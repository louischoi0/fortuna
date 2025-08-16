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
	mu      sync.Mutex
	SpaceID string

	CurrentPage *Page
	LastPage    *Page
	LastPageNum uint64
	Pages       []*Page

	Storage *storage.FileStorage
	meta    *grocksdb.DB

	BlockNum    uint64
	CurrentFile *storage.File

	LastAppendedAt time.Time
}

func NewSpace(spaceID string) *Space {
	dir := filepath.Join(storage.DataRootDir(), fmt.Sprintf("space.%s", spaceID))
	storage := storage.NewFileStorage(dir)
	metaDB, err := rock.GetDBInstance(fmt.Sprintf("space_additional.%s", spaceID))

	if err != nil {
		log.Fatalf("failed to get space meta db: %v", err.Error())
	}

	page := NewPage(spaceID, 1, nil)

	return &Space{
		CurrentPage: page,
		LastPageNum: 0,
		LastPage:    nil,
		SpaceID:     spaceID,
		meta:        metaDB,
		Storage:     storage,
	}
}

func (s *Space) LoadSpaceData() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	pageNum := s.ReadLastPageNum()
	s.LastPageNum = pageNum

	log.Printf("loading space data for %s, last page num: %v", s.SpaceID, s.LastPageNum)

	if len(s.Pages) != 0 {
		return fmt.Errorf("space loaded again")
	}

	for i := uint64(1); i <= s.LastPageNum; i++ {
		blk, err := s.ReadPage(i)
		if err != nil {
			return err
		}
		s.Pages = append(s.Pages, blk)
	}

	if s.LastPageNum > 0 {
		s.LastPage = s.Pages[s.LastPageNum-1]
	}

	return nil
}

func GetFileName(blockNum uint64) string {
	return fmt.Sprintf(C.PAGE_FILE_NAME_FORMAT, blockNum)
}

func (s *Space) NewNextPageFile() error {
	s.BlockNum++
	fileName := GetFileName(s.BlockNum)
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

func (s *Space) CommitPage(page *Page) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if pn := s.ReadLastPageNum(); pn != s.LastPageNum {
		return fmt.Errorf("space last page num mismatch expected %v, but %v", s.LastPageNum, pn)
	}

	if s.LastPageNum+1 != page.GetPageNum() {
		return fmt.Errorf("last page num mismatch expected %v, but %v", s.LastPageNum+1, page.GetPageNum())
	}

	if page.GetPageNum() != 1 && s.LastPage != nil && s.LastPage.Hash() != page.PrevPageHash {
		return fmt.Errorf("previous block hash mismatch expected %s not %s", s.LastPage.Hash(), page.PrevPageHash)
	}
	if page.GetPageNum() == 1 && s.LastPage != nil {
		return fmt.Errorf("first block must be genesis block")
	}
	if page.GetPageNum() > 1 && s.LastPage == nil {
		return fmt.Errorf("last block is nil")
	}

	if s.CurrentFile == nil || s.IsFileSizeExceed(s.CurrentFile) {
		s.NewNextPageFile()
	}

	page_buffer, err := page.Encode()
	if err != nil {
		return err
	}

	pageIndex := NewPageIndex(s.SpaceID, page.GetPageNum(), s.BlockNum, uint64(s.CurrentFile.GetSize()), uint64(len(page_buffer)))
	err = s.WritePageIndex(pageIndex)
	if err != nil {
		return err
	}

	err = s.CurrentFile.AppendFileBytes(page_buffer)
	if err != nil {
		return err
	}

	err = s.SetPageNum(page.GetPageNum())
	if err != nil {
		return err
	}

	s.LastPage = s.CurrentPage
	s.LastPageNum++
	s.CurrentPage = page
	s.LastAppendedAt = time.Now()

	return nil
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

func (s *Space) ReadPage(num uint64) (*Page, error) {
	pageIndex, err := s.ReadPageIndex(num)
	if err != nil {
		return nil, fmt.Errorf("failed to get PageIndex: %v", err.Error())
	}

	blockNum := pageIndex.BlockNum
	fileName := GetFileName(blockNum)
	file := s.Storage.GetOrOpenFile(fileName)

	fmt.Printf("offset: %v size: %v", pageIndex.Offset, pageIndex.Size)

	pageBytes, err := file.OffsetRead(int64(pageIndex.Offset), int64(pageIndex.Size))
	if err != nil {
		return nil, err
	}

	page, err := DecodePage(pageBytes)
	if err != nil {
		return nil, err
	}

	return page, nil
}

func (s *Space) ReadPageIndex(pageNum uint64) (*PageIndex, error) {
	key := fmt.Sprintf(C.PAGE_INDEX_KEY_FORMAT, pageNum)
	value, err := rock.GetValue(s.meta, key)
	if err != nil {
		return nil, err
	}

	pageIndex, err := DecodePageIndex(value)
	if err != nil {
		return nil, err
	}

	return pageIndex, nil
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
	s.LastPageNum = pageNum
	return nil
}

func (s *Space) ReadLastPageNum() uint64 {
	pageNum, err := rock.GetValue(s.meta, "page_num")
	if err != nil {
		return 0
	}

	pn, err := util.DecodeUint64(pageNum)
	if err != nil {
		return 0
	}
	return pn
}

func (s *Space) GetPageNum() uint64 {
	return s.LastPageNum
}

func (s *Space) Next() uint64 {
	s.mu.Lock()
	defer s.mu.Unlock()

	log.Printf("next num: %v", s.LastPageNum+1)

	return s.LastPageNum + 1
}

func (s *Space) SetMetaDB(meta *grocksdb.DB) {
	s.meta = meta
}
