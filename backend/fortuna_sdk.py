import os
import subprocess

BIN_ALIAS = "fortuna"
BIN = os.path.join("/Users/louis/fortuna", BIN_ALIAS)

def execute_sdk(apiType, apiName, *args):
    print(f"Executing {BIN} {apiType} {apiName} {args}")
    result = subprocess.run([BIN, apiType, apiName, *args], capture_output=True, text=True)
    return result.stdout

if __name__ == "__main__":
    execute_sdk("oracle", "status")