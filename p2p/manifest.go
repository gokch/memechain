package p2p

import (
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"sort"
	"time"

	"github.com/gokch/memechain/types"
)

type Manifest struct {
	// p2pcp 의 version, 같은 major version 끼리는 통신 보장
	MajorVersion string `json:"majorVersion"`
	MinorVersion string `json:"minorVersion"`

	// 전송할 file 목록으로 chunk 생성시 사용할 크기
	ChunkSize int64 `json:"chunkSize"`

	// chunk 목록에 대한 sha-1 checksum
	// 각 p2pcp 는 file 목록으로 chunk 목록을 생성하고 이와 비교해서 동일한 chunk 목록을 갖게함
	ChunkChecksum      string `json:"checksum"`
	TransferChunkCount int    `json:"transferChunkCount"`

	Dirs       []types.Dir       `json:"dirs"`
	Symlinks   []types.Symlink   `json:"symlinks"`
	EmptyFiles []types.EmptyFile `json:"emptyFiles"`
	Files      []types.File      `json:"files"`

	// Manifest 의 Marshal 결과를 재사용 하기 위함
	JsonData []byte `json:"-"`
}

func (m *Manifest) Init(dirSrc string, peerList []string) ([]types.TransferChunk, error) {
	m.MajorVersion = MajorVersion
	m.MinorVersion = MinorVersion
	m.ChunkSize = *flagChunkSize

	var chunks []types.TransferChunk
	var err error

	if dirSrc != "" {
		DebugPrintf("Making local files list...")

		chunks, err = m.scanDirectoryLocal(dirSrc)
		if err != nil {
			ErrorPrintf("Failed to make local files list.: %v", err)
			return nil, err
		}
	} else if len(peerList) > 0 {
		DebugPrintf("Waiting peer who has complete files list...")

		chunks, err = m.scanDirectoryPeer(peerList)
		if err != nil {
			ErrorPrintf("Failed to make remote files list.: %v", err)
			return nil, err
		}
	} else {
		return nil, errors.New("at least one of local source directory or peer list required")
	}

	m.JsonData, err = json.Marshal(mainContext.Manifest)
	if err != nil {
		ErrorPrintf("manifest JSON Marshal failed.:%v", err)
		return nil, err
	}

	return chunks, nil
}

func (m *Manifest) createTransferChunks(files []types.File) []types.TransferChunk {
	transferChunkList := make([]types.TransferChunk, 0)
	if len(files) == 0 {
		return transferChunkList
	}

	transferChunk := types.TransferChunk{}
	var transferChunkSize int64

	newTransferChunk := func() {
		transferChunkList = append(transferChunkList, transferChunk)
		transferChunk = types.TransferChunk{}
		transferChunkSize = 0
	}
	addToTransferChunk := func(index int, from int64, to int64) {
		transferChunk.FileChunks = append(transferChunk.FileChunks, types.FileChunk{index, from, to})
		transferChunkSize += (to - from)
	}

	for index := range files {
		file := &files[index]

		var from int64
		remainFilesize := file.Size
		for remainFilesize > 0 {
			if transferChunkSize+remainFilesize < m.ChunkSize {
				// TransferChunk 에 남은 file 을 담을수 있음
				addToTransferChunk(index, from, file.Size)

				if len(transferChunk.FileChunks) == *flagMaxFileCountPerChunk {
					// transferChunk.FileChunks 길이가 flagMaxFileCountPerChunk 에 도달하여 TransferChunk 를 분리함
					newTransferChunk()
				}

				// remainFilesize = 0
				break
			} else {
				// TransferChunk 에 남은 file 을 담을수 없어 자름
				size := m.ChunkSize - transferChunkSize
				addToTransferChunk(index, from, from+size)
				newTransferChunk()

				from += size
				remainFilesize -= size
			}
		}
	}

	if len(transferChunk.FileChunks) > 0 {
		transferChunkList = append(transferChunkList, transferChunk)
	}

	return transferChunkList
}

func (m *Manifest) createChunkChecksum(files []types.File, transferChunkList []types.TransferChunk) (string, error) {
	hash := sha1.New()

	jsonData, err := json.Marshal(files)
	if err != nil {
		ErrorPrintf("JSON Marshal failed.:%v", err)
		return "", err
	}
	hash.Write(jsonData)

	jsonData, err = json.Marshal(transferChunkList)
	if err != nil {
		ErrorPrintf("JSON Marshal failed.:%v", err)
		return "", err
	}
	hash.Write(jsonData)

	return hex.EncodeToString(hash.Sum(nil)), nil
}

func (m *Manifest) scanDirectoryEachPeer(peerHost string) (chunks []types.TransferChunk, err error) {
	url := fmt.Sprintf("http://%v/manifest", peerHost)
	options := NewHttpRequestOptions()
	options.EncodingType = *flagCompressType
	options.ExpectedContentType = "application/json"

	var content []byte
	content, err = RequestHttp(url, options)
	if err != nil {
		return
	}

	var resultObject Manifest
	err = json.Unmarshal(content, &resultObject)
	if err != nil {
		return
	}

	if resultObject.MajorVersion != m.MajorVersion {
		err = fmt.Errorf("peer %v has different major version.:expected=%v.%v, actual=%v.%v", peerHost, m.MajorVersion, m.MajorVersion, resultObject.MajorVersion, resultObject.MajorVersion)
		return
	}

	if resultObject.ChunkSize != m.ChunkSize {
		err = fmt.Errorf("peer %v has different chunk size.:expected=%v, actual=%v", peerHost, m.ChunkSize, resultObject.ChunkSize)
		return
	}

	sort.Slice(resultObject.Files, func(i, j int) bool {
		return resultObject.Files[i].Name < resultObject.Files[j].Name
	})
	chunks = m.createTransferChunks(resultObject.Files)
	chunkChecksum, err := m.createChunkChecksum(resultObject.Files, chunks)
	if err != nil {
		return
	}
	transferChunkCount := len(chunks)

	if resultObject.ChunkChecksum != chunkChecksum {
		err = fmt.Errorf("peer %v has different chunk checksum.:expected=%v, actual=%v", peerHost, chunkChecksum, resultObject.ChunkChecksum)
		return
	}

	if resultObject.TransferChunkCount != transferChunkCount {
		err = fmt.Errorf("peer %v has different chunk count.:expected=%v, actual=%v", peerHost, transferChunkCount, resultObject.TransferChunkCount)
		return
	}

	m.Dirs = resultObject.Dirs
	m.Symlinks = resultObject.Symlinks
	m.EmptyFiles = resultObject.EmptyFiles
	m.Files = resultObject.Files
	m.ChunkChecksum = resultObject.ChunkChecksum
	m.TransferChunkCount = resultObject.TransferChunkCount

	return
}

func (m *Manifest) scanDirectoryPeer(peerList []string) ([]types.TransferChunk, error) {
	t0 := time.Now()
	for {
		tl0 := time.Now()
		for _, peer := range peerList {
			chunks, err := m.scanDirectoryEachPeer(peer)
			if err == nil {
				return chunks, nil
			}
			DebugPrintf("failed to getting files list from peer %v: %v", peer, err)
		}
		tl1 := time.Now()
		if tl1.Sub(t0) > *flagPeerWaitTimeout {
			break
		}

		elapsed := tl1.Sub(tl0)
		if elapsed < InitScanPeerInterval {
			time.Sleep(InitScanPeerInterval - elapsed)
		}
	}

	return nil, errors.New("no peer available")
}

func (m *Manifest) scanDirectoryLocalTarget(dirSrc string, subDir string, dirs *[]types.Dir, symlinks *[]types.Symlink, emptyFiles *[]types.EmptyFile, files *[]types.File) error {
	target := dirSrc + subDir
	localFiles, err := os.ReadDir(target)
	if err != nil {
		return err
	}

	for _, localFile := range localFiles {
		name := subDir + localFile.Name()
		info, err := localFile.Info()
		if err != nil {
			ErrorPrintf("%v: get info failed:%s", name, err.Error())
			return err
		}

		mode := info.Mode()
		if mode.IsDir() {
			err := m.scanDirectoryLocalTarget(dirSrc, name+"/", dirs, symlinks, emptyFiles, files)
			if err != nil {
				return err
			}
			*dirs = append(*dirs, types.Dir{Name: name, Perm: mode.Perm()})
		} else if mode&os.ModeSymlink != 0 {
			symlinkTo, err := os.Readlink(dirSrc + name)
			if err != nil {
				ErrorPrintf("%v: Readlink failed:%s", name, err.Error())
				return err
			}
			*symlinks = append(*symlinks, types.Symlink{Name: name, SymlinkTo: symlinkTo})
		} else if mode.IsRegular() {
			size := info.Size()
			if size == 0 {
				*emptyFiles = append(*emptyFiles, types.EmptyFile{Name: name, Perm: mode.Perm()})
			} else {
				*files = append(*files, types.File{Name: name, Size: size, Perm: mode.Perm()})
			}
		} else {
			ErrorPrintf("%v: unsupported mode(%v). ignored.", name, mode)
			continue
		}
	}

	return nil
}

func (m *Manifest) scanDirectoryLocal(dirSrc string) ([]types.TransferChunk, error) {
	dirs := make([]types.Dir, 0)
	symlinks := make([]types.Symlink, 0)
	emptyFiles := make([]types.EmptyFile, 0)
	files := make([]types.File, 0)
	err := m.scanDirectoryLocalTarget(dirSrc, "/", &dirs, &symlinks, &emptyFiles, &files)
	if err != nil {
		return nil, err
	}

	sort.Slice(files, func(i, j int) bool {
		return files[i].Name < files[j].Name
	})
	chunks := m.createTransferChunks(files)
	chunkChecksum, err := m.createChunkChecksum(files, chunks)
	if err != nil {
		return nil, err
	}
	transferChunkCount := len(chunks)

	m.Dirs = dirs
	m.Symlinks = symlinks
	m.EmptyFiles = emptyFiles
	m.Files = files
	m.ChunkChecksum = chunkChecksum
	m.TransferChunkCount = transferChunkCount

	return chunks, nil
}
