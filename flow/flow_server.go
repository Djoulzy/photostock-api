package flow

import (
	"archive/zip"
	"bytes"
	"io"
	"io/fs"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strconv"
	"strings"

	"github.com/Djoulzy/IStock2/database"
	"github.com/Djoulzy/IStock2/diskcopy"
	"github.com/Djoulzy/IStock2/model"
)

var (
	completedFiles = make(chan string, 100)
	conf           *model.ConfigData
	DB             *database.Node
)

type ByChunk []fs.DirEntry

func (a ByChunk) Len() int      { return len(a) }
func (a ByChunk) Swap(i, j int) { a[i], a[j] = a[j], a[i] }
func (a ByChunk) Less(i, j int) bool {
	ai, _ := strconv.Atoi(a[i].Name())
	aj, _ := strconv.Atoi(a[j].Name())
	return ai < aj
}

func StartUploadFileAssembler(db *database.Node, config *model.ConfigData) {
	conf = config
	DB = db
	runtime.GOMAXPROCS(runtime.NumCPU())
	for range 8 {
		go assembleFile(completedFiles)
	}
}

func ContinueUpload(r *http.Request) error {
	chunkDirPath := conf.ChunkDir + "/" + r.FormValue("flowFilename") + "/" + r.FormValue("flowChunkNumber")
	if _, err := os.Stat(chunkDirPath); err != nil {
		log.Println("Chunk not found: ", chunkDirPath)
		return err
	}
	return nil
}

func StreamingReader(r *http.Request) error {
	buf := new(bytes.Buffer)
	reader, err := r.MultipartReader()
	// Part 1: Chunk Number
	// Part 4: Total Size (bytes)
	// Part 6: File Name
	// Part 8: Total Chunks
	// Part 9: Chunk Data
	if err != nil {
		return err
	}

	part, err := reader.NextPart() // 1
	if err != nil {
		return err
	}
	io.Copy(buf, part)
	chunkNo := buf.String()
	buf.Reset()
	log.Println("Chunk No: ", chunkNo)

	for range 3 { // 2 3 4
		// move through unused parts
		part, err = reader.NextPart()
		if err != nil {
			return err
		}
	}

	io.Copy(buf, part)
	flowTotalSize := buf.String()
	buf.Reset()

	for range 2 { // 5 6
		// move through unused parts
		part, err = reader.NextPart()
		if err != nil {
			return err
		}
	}

	io.Copy(buf, part)
	fileName := buf.String()
	buf.Reset()

	for range 3 { // 7 8 9
		// move through unused parts
		part, err = reader.NextPart()
		if err != nil {
			return err
		}
	}

	chunkDirPath := conf.ChunkDir + "/" + fileName
	err = os.MkdirAll(chunkDirPath, 02750)
	if err != nil {
		return err
	}
	log.Println("Chunk Dir: ", chunkDirPath)

	dst, err := os.Create(chunkDirPath + "/" + chunkNo)
	if err != nil {
		return err
	}
	log.Println("Chunk File: ", chunkDirPath+"/"+chunkNo)
	defer dst.Close()
	io.Copy(dst, part)
	log.Println("Chunk Size: ", flowTotalSize)

	fileInfos, err := os.ReadDir(chunkDirPath)
	if err != nil {
		return err
	}
	log.Println("Total Size: ", flowTotalSize)

	if flowTotalSize == strconv.Itoa(int(totalSize(fileInfos))) {
		completedFiles <- chunkDirPath
	}
	return nil
}

func totalSize(fileInfos []fs.DirEntry) int64 {
	var sum int64
	for _, fi := range fileInfos {
		info, err := fi.Info()
		if err != nil {
			return 0
		}
		sum += info.Size()
	}
	return sum
}

func assembleFile(jobs <-chan string) {
	for path := range jobs {
		galName, err := assemble(path)
		if err == nil {
			err := diskcopy.LookupNewDir(DB, conf.ImportDir, conf.AbsoluteBankPath)
			if err != nil {
				log.Println("Error looking up new directory: ", err)
				os.RemoveAll(galName)
			}
		}
	}
}

func assemble(path string) (galName string, err error) {
	log.Println("Assembling: ", path)
	fileInfos, err := os.ReadDir(path)
	if err != nil {
		return "", err
	}

	// create final file to write to
	destPath := conf.ImportDir + "/" + filepath.Base(path)
	log.Println("Final File: ", destPath)
	dst, err := os.Create(destPath)
	if err != nil {
		return
	}
	defer dst.Close()

	sort.Sort(ByChunk(fileInfos))
	for _, fs := range fileInfos {
		src, err := os.Open(path + "/" + fs.Name())
		if err != nil {
			return "", err
		}
		defer src.Close()
		io.Copy(dst, src)
	}

	if filepath.Ext(destPath) == ".zip" {
		log.Println("Decompressing: ", destPath)
		galName, err = extractZip(destPath)
		if err != nil {
			log.Println("Error decompressing: ", err)
			return "", err
		}
		os.Remove(destPath)
	}
	os.RemoveAll(path)
	return galName, nil
}

func extractZip(path string) (galName string, err error) {
	basePath := filepath.Dir(path)
	// Open a zip archive for reading.
	log.Println("Extracting zip file:", path)
	r, err := zip.OpenReader(path)
	if err != nil {
		return "", err
	}
	defer r.Close()

	// Iterate through the files in the archive,
	// printing some of their contents.
	if err != nil {
		return "", err
	}
	for _, f := range r.File {
		log.Println("Extracting file:", f.Name)
		baseFileName := filepath.Base(f.Name)

		// Exclude hidden files
		if strings.HasPrefix(baseFileName, ".") {
			// log.Println("Drop file:", baseFileName)
			continue
		}
		newFilePath := basePath + "/" + f.Name
		if f.FileInfo().IsDir() {
			// Create a directory on the filesystem.
			if err = os.MkdirAll(newFilePath, os.ModePerm); err != nil {
				return "", err
			}
			galName = newFilePath
			continue
		}

		// Open the file inside the zip archive.
		rc, err := f.Open()
		if err != nil {
			return "", err
		}
		defer rc.Close()

		// Create a file on the filesystem.
		outFile, err := os.Create(newFilePath)
		if err != nil {
			return "", err
		}

		// Copy the contents of the file from the zip archive to the filesystem.
		if _, err = io.Copy(outFile, rc); err != nil {
			return "", err
		}
		outFile.Close()
	}
	// Delete the zip file after extraction
	if err = os.Remove(path); err != nil {
		return "", err
	}
	return galName, nil
}
