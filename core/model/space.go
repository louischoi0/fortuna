package model

import (
	"bytes"
	"fmt"
	"fortuna/core/storage"
	"fortuna/rock"
	"fortuna/util"
	"log"
	"path/filepath"
	"sync"
	"time"

	C "fortuna/core/config"

	"github.com/linxGnu/grocksdb"
)

var __spaces map[string]*Space

type PageIndex struct {
	SpaceID  string
	PageNum  uint64
	BlockNum uint64
	Offset   uint64
	Size     uint64
}

func (b *PageIndex) Encode() ([]byte, error) {
	var buffer bytes.Buffer
	buffer.WriteString(b.SpaceID)

	buffer.Write(util.EncodeUint64(uint64(b.PageNum)))
	buffer.Write(util.EncodeUint64(uint64(b.BlockNum)))
	buffer.Write(util.EncodeUint64(uint64(b.Offset)))
	buffer.Write(util.EncodeUint64(uint64(b.Size)))

	return buffer.Bytes(), nil
}

func DecodePageIndex(data []byte) (*PageIndex, error) {
	b := &PageIndex{}

	var (
		off = 0
		n   = len(data)
	)

	need := func(k int) error {
		if off+k > n {
			return fmt.Errorf("buffer underflow: need %d bytes, value: %s", k, string(data))
		}
		return nil
	}
	readFixedString := func(k int) (string, error) {
		if err := need(k); err != nil {
			return "", err
		}
		s := string(data[off : off+k])
		off += k
		return s, nil
	}
	readU64LE := func() (uint64, error) {
		if err := need(8); err != nil {
			return 0, err
		}
		v, err := util.DecodeUint64(data[off : off+8])
		if err != nil {
			return 0, err
		}
		off += 8
		return v, nil
	}

	spaceID, err := readFixedString(64)
	if err != nil {
		return nil, err
	}
	b.SpaceID = spaceID

	pageNum, err := readU64LE()
	if err != nil {
		return nil, err
	}
	b.PageNum = pageNum

	blockNum, err := readU64LE()
	if err != nil {
		return nil, err
	}
	b.BlockNum = blockNum

	offset, err := readU64LE()
	if err != nil {
		return nil, err
	}
	b.Offset = offset

	size, err := readU64LE()
	if err != nil {
		return nil, err
	}
	b.Size = size

	return b, nil
}

func NewPageIndex(spaceID string, pageNum uint64, blockNum uint64, offset uint64, size uint64) *PageIndex {
	if len(spaceID) != 64 {
		log.Fatalf("spaceID has invalid length %s", spaceID)
	}

	return &PageIndex{
		SpaceID:  spaceID,
		PageNum:  pageNum,
		BlockNum: blockNum,
		Offset:   offset,
		Size:     size,
	}
}

type Space struct {
	mu sync.Mutex

	SpaceID     string
	CurrentPage *Page

	LastPage             	*Page
	LastCommittedPageNum 	uint64
	ActivePageNum		uint64

	Pages       []*Page
	Storage     *storage.FileStorage
	meta        *grocksdb.DB
	FileNum     uint64
	CurrentFile *storage.File

	LastAppendedAt time.Time
	IsReplica      bool
}

func NewSpace(spaceID string, metaDB *grocksdb.DB) *Space {

	dir := filepath.Join(storage.DataRootDir(), fmt.Sprintf("space.%s", spaceID))
	storage := storage.NewFileStorage(dir)

	if _, ok := __spaces[spaceID]; ok {
		log.Fatalf("space %s initialized again, not allowed", spaceID)
	}

	pageNum, err := ReadLastPageNum(metaDB)
	if err != nil {
		log.Fatalf("failed to get last page num: %v", err.Error())
	}

	var space *Space

	if pageNum != 0 {
		space = &Space{
			SpaceID: spaceID,
			Storage: storage,
			meta:    metaDB,
		}
		space.LoadSpaceData()
		space.CommitSpaceHeader()
		log.Printf("space %s loaded, pagenum: %v", spaceID, space.LastCommittedPageNum)
		return space
	}

	page := NewPage(spaceID, 1, nil)

	space = &Space{
		CurrentPage: page,
		LastPage:    nil,
		SpaceID:     spaceID,
		meta:        metaDB,
		Storage:     storage,
		IsReplica:   false,
		LastCommittedPageNum: 0,
		ActivePageNum: 1,
	}

	log.Printf("initailized space %s, pagenum: %v", spaceID, space.LastCommittedPageNum)
	space.CommitSpaceHeader()
	return space
}

func (s *Space) LoadSpaceData() error {
	pageNum := s.ReadLastPageNum()

	log.Printf("loading space data for %s, last page num: %v", s.SpaceID, pageNum)

	if len(s.Pages) != 0 {
		return fmt.Errorf("space loaded again")
	}

	for i := uint64(1); i <= pageNum; i++ {
		blk, err := s.ReadPage(i)
		if err != nil {
			log.Fatalf("failed to read page %v: %v", i, err.Error())
		}
		s.Pages = append(s.Pages, blk)
	}

	if pageNum > 0 {
		s.LastPage = s.Pages[pageNum-1]
	}

	s.LastCommittedPageNum = pageNum
	s.CurrentPage = NewPage(s.SpaceID, pageNum+1, s.LastPage)
	s.ActivePageNum = pageNum + 1

	return nil
}

func GetFileName(blockNum uint64) string {
	return fmt.Sprintf(C.PAGE_FILE_NAME_FORMAT, blockNum)
}

func (s *Space) NewNextPageFile() error {
	s.FileNum++
	fileName := GetFileName(s.FileNum)
	s.CurrentFile = storage.NewFile(s.Storage, fileName)

	err := s.SetCurrentFile(s.CurrentFile)
	if err != nil {
		return err
	}

	return nil
}

func (s *Space) IsFileSizeExceed(f *storage.File) bool {
	return f.GetSize() >= C.PAGE_FILE_SIZE
}

func (s *Space) MaybeCommitPage() (*Page, error) {

	if s.CurrentPage == nil {
		log.Fatalf("space %s current page is nil", s.SpaceID)
	}

	if s.CurrentPage.GetCount() < C.PAGE_MIN_EVENT_COUNT {
		return nil, nil
	}

	return s.CommitCurrentPage()
}

func (s *Space) CommitCurrentPage() (*Page, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	page := s.CurrentPage
	page.Update()

	log.Printf("start to commit current page: %s", s.SpaceID)

	if s.ReadLastPageNum() != page.GetPageNum()-1 {
		return nil, fmt.Errorf("space last page num mismatch expected %v, but %v", page.GetPageNum()-1, s.ReadLastPageNum())
	}

	if page.GetPageNum() != 1 && s.LastPage != nil && s.LastPage.Hash() != page.PrevPageHash {
		return nil, fmt.Errorf("previous block hash mismatch expected %s not %s", s.LastPage.Hash(), page.PrevPageHash)
	}
	if page.GetPageNum() == 1 && s.LastPage != nil {
		return nil, fmt.Errorf("first block must be genesis block")
	}
	if page.GetPageNum() > 1 && s.LastPage == nil {
		return nil, fmt.Errorf("last block is nil")
	}

	if s.LastPage != nil && s.LastPage.GetPageNum() != page.GetPageNum()-1 {
		return nil, fmt.Errorf("last page num mismatch expected %v, but %v", s.LastPage.GetPageNum(), page.GetPageNum()-1)
	}

	if s.CurrentFile == nil || s.IsFileSizeExceed(s.CurrentFile) {
		s.NewNextPageFile()
	}

	page_buffer, err := page.Encode()
	if err != nil {
		return nil, err
	}

	pageIndex := NewPageIndex(s.SpaceID, page.GetPageNum(), s.FileNum, uint64(s.CurrentFile.GetSize()), uint64(len(page_buffer)))
	err = s.WritePageIndex(pageIndex)
	if err != nil {
		return nil, err
	}

	err = s.CurrentFile.AppendFileBytes(page_buffer)
	if err != nil {
		return nil, err
	}

	err = s.SetPageNum(page.GetPageNum())
	if err != nil {
		return nil, err
	}

	s.LastPage = page
	page.Committed = true

	s.LastCommittedPageNum = page.GetPageNum()

	s.CurrentPage = NewPage(s.SpaceID, page.GetPageNum()+1, s.LastPage)
	s.LastAppendedAt = time.Now()

	return page, nil
}

func (s *Space) GetPage(num uint64) (*Page, error) {

	if int(num) > len(s.Pages)+1 {
		return nil, fmt.Errorf("page not found for %v", num)
	}

	page := s.Pages[num-1]

	if page.GetPageNum() != num {
		log.Fatalf("page invalid error page expected num: %v, but %v", page.GetPageNum(), num)
	}

	return page, nil
}

func (s *Space) Reset() error {
	if err := s.Storage.DeleteAll(); err != nil {
		return err
	}
	if err := rock.ClearDB(s.meta); err != nil {
		return err
	}

	return nil
}

func ReadPage(db *grocksdb.DB, storage *storage.FileStorage, num uint64) (*Page, error) {
	pageIndex, err := ReadPageIndex(db, num)
	if err != nil {
		return nil, fmt.Errorf("failed to get PageIndex: %v", err.Error())
	}

	blockNum := pageIndex.BlockNum
	fileName := GetFileName(blockNum)
	file := storage.GetOrOpenFile(fileName)

	log.Printf("read page %v offset: %v size: %v", num, pageIndex.Offset, pageIndex.Size)

	pageBytes, err := file.OffsetRead(int64(pageIndex.Offset), int64(pageIndex.Size))
	if err != nil {
		return nil, err
	}

	page, err := DecodePage(pageBytes)
	if err != nil {
		return nil, err
	}

	page.Update()
	return page, nil
}

func (s *Space) ReadPage(num uint64) (*Page, error) {
	return ReadPage(s.meta, s.Storage, num)
}

func ReadPageIndex(db *grocksdb.DB, num uint64) (*PageIndex, error) {
	key := fmt.Sprintf(C.PAGE_INDEX_KEY_FORMAT, num)
	value, err := rock.GetValue(db, key)
	if err != nil {
		return nil, err
	}

	pageIndex, err := DecodePageIndex(value)
	if err != nil {
		return nil, err
	}

	return pageIndex, nil
}

func (s *Space) ReadPageIndex(pageNum uint64) (*PageIndex, error) {
	return ReadPageIndex(s.meta, pageNum)
}

func (s *Space) WritePageIndex(pageIndex *PageIndex) error {
	key := fmt.Sprintf(C.PAGE_INDEX_KEY_FORMAT, pageIndex.PageNum)
	pageIndexBytes, err := pageIndex.Encode()
	if err != nil {
		return err
	}

	rock.SetValue(s.meta, key, pageIndexBytes)
	log.Printf("write page index at num %v - size:%v offset: %v buffer_size: %v", pageIndex.PageNum, pageIndex.Size, pageIndex.Offset, len(pageIndexBytes))
	return nil
}

func (s *Space) SetCurrentFile(file *storage.File) error {
	err := rock.SetValue(s.meta, "current", []byte(file.Name))
	if err != nil {
		return err
	}

	s.CurrentFile = file
	return nil
}

func (s *Space) SetPageNum(pageNum uint64) error {
	err := rock.SetValue(s.meta, "page_num", []byte(util.EncodeUint64(pageNum)))
	if err != nil {
		return err
	}
	return nil
}

func ReadLastPageNum(db *grocksdb.DB) (uint64, error) {
	pageNum, err := rock.GetValue(db, "page_num")
	if err != nil {
		return 0, err
	}

	if len(pageNum) == 0 {
		return 0, nil
	}

	pn, err := util.DecodeUint64(pageNum)
	if err != nil {
		return 0, err
	}

	return pn, nil
}

func (s *Space) ReadLastPageNum() uint64 {
	pageNum, err := ReadLastPageNum(s.meta)
	if err != nil {
		return 0
	}

	return pageNum
}

func (s *Space) SetMetaDB(meta *grocksdb.DB) {
	s.meta = meta
}

type SpaceInfo struct {
	Hash                 string `json:"hash"`
	ActivePageNum        int64  `json:"active_page_num"`
	LastCommittedPageNum int64  `json:"last_committed_page_num"`
}

func (s *Space) Hash() string {
	return "thisishash"
}

func (s *Space) Info() SpaceInfo {
	return SpaceInfo{
		Hash:                 s.Hash(),
		ActivePageNum:        int64(s.ActivePageNum),
		LastCommittedPageNum: int64(s.LastCommittedPageNum),
	}
}

func (s *Space) CommitSpaceHeader() error {
	log.Printf("commiting space header for %s", s.SpaceID)

	err := rock.SetValue(s.meta, fmt.Sprintf("space-%v", s.SpaceID), util.EncodeUint64(s.LastCommittedPageNum))
	if err != nil {
		log.Fatalf("failed to set space id: %v", err.Error())
	}

	return nil
}

func ListSpaces(db *grocksdb.DB) ([]string, error) {
	cb := func(key string, value []byte) string {
		return key
	}

	return rock.ScanC(db, "space-", cb)
}
