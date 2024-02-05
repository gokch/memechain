package p2p

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"math/rand"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"sync/atomic"
	"syscall"
	"time"

	"github.com/gokch/memechain/types"
	"github.com/gokch/memechain/utilx"

	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
)

/* 빌드할 때 변수가 설정됨. Makefile 참고. */
var MajorVersion string
var MinorVersion string
var PatchVersion string
var BuildVersion string

var mainContext MainContext

/*
* peer 연결에 대한 timeout 설정으로 다음의 상황에서 사용
  - peer-list-url 에 설정된 peer 목록 획득
  - peer 목록에서 seeder 를 찾고 chunk 목록 획득
  - chunk 목록 동기화 및 다운로드
*/
const PeerTransferTimeout = 20 * time.Second
const DurationDownloaderSleep = 200 * time.Millisecond
const InitScanPeerInterval = 500 * time.Millisecond
const ShutdownTimeout = 10 * time.Second
const MaxPeerNumConcurrent = 16
const MaxLocalNumConcurrent = 16
const MaxChunkSize = 134217728
const HttpConnectTimeout = 500 * time.Millisecond

const (
	EncodingTypeNone = "none"
	EncodingTypeZstd = "zstd"
)

var (
	flagVersion                   = flag.Bool("version", false, "Show version information.")
	flagVerbose                   = flag.Bool("verbose", false, "Log verbose information.")
	flagStatistics                = flag.Bool("statistics", false, "Print statistics information on complete.")
	flagWaitHost                  = flag.String("wait-host", "", "The host to just wait for the completion of transfer. Ignores other options except -wait-host-timeout.")
	flagWaitHostTimeout           = flag.Duration("wait-host-timeout", 10*time.Minute, "Timeout for -wait-host. (0=wait infinitely)")
	flagExitComplete              = flag.Bool("exit-complete", false, "After completed download, exit immediately.")
	flagPeerNumConcurrent         = flag.Int("peer-num-concurrent", 2, fmt.Sprintf("Number of concurrent download jobs per peer.(1-%v)", MaxPeerNumConcurrent))
	flagLocalNumConcurrent        = flag.Int("local-num-concurrent", 1, fmt.Sprintf("Number of concurrent transfer jobs for local.(1-%v)", MaxLocalNumConcurrent))
	flagListenAddr                = flag.String("listen-addr", "0.0.0.0:10090", "TCP Address for listen.")
	flagCompressType              = flag.String("compress-type", "none", "Compress type for network data transfer.(none|zstd)")
	flagChunkSize                 = flag.Int64("chunk-size", 16777216, "Size of transfer unit.")
	flagMaxFileCountPerChunk      = flag.Int("files-per-chunk", 5, "Max count of files per chunk.")
	flagPeerList                  = flag.String("peer-list", "", "Comma separated peer host list.")
	flagPeerListURL               = flag.String("peer-list-url", "", "HTTP URL which contains comma separated peer host list.")
	flagPeerWaitTimeout           = flag.Duration("peer-wait-timeout", 20*time.Second, "Timeout for wait until first peer becomes online.")
	flagTransferIdleTimeout       = flag.Duration("transfer-idle-timeout", 30*time.Second, "Timeout for transfer idle to treat to failure. (0=wait infinitely)")
	flagAvailableListSyncInterval = flag.Duration("sync-interval", 2000*time.Millisecond, "Interval to synchronize the list of available chunks between peers.")
	flagSrc                       = flag.String("src", "", "Source directory. Required.")
	flagDst                       = flag.String("dst", "", "Destination directory. If not specified. It will perform serve-only mode.")
	flagShowHelp                  = flag.Bool("help", false, "Show usage")
)

type MainContext struct {
	Hostname           string
	UUID               string
	PeerList           []string
	PeerNumConcurrent  int
	LocalNumConcurrent int

	Manifest       Manifest
	TransferChunks []types.TransferChunk

	WantStopDownload atomic.Bool
	Completed        atomic.Bool
	LastUpdatedAt    utilx.AtomicTime
	IdleTimeout      time.Duration

	WGDownloader sync.WaitGroup
	ServeOnly    bool

	DirSrc string
	DirDst string

	/* 통계정보 */
	PrintStatistics         bool
	MutexStatistics         sync.Mutex
	Statistics              map[string]string
	BytesReceivedTotal      atomic.Int64
	BytesTransmittedTotal   atomic.Int64
	ChunksReceivedTotal     atomic.Int64
	CumulativeDurationTotal utilx.AtomicDuration
}

func peerListFromURL(mainContext *MainContext, url string, timeout time.Duration) ([]string, error) {
	if !strings.HasPrefix(url, "http://") && !strings.HasPrefix(url, "https://") {
		return nil, fmt.Errorf("peer-list-url scheme not allowed.:%v", url)
	}
	options := NewHttpRequestOptions()

	/* 외부서버에서의 압축 응답이 p2pcp에서 완전히 지원가능한 상태가 아니기 때문에 압축 요청을 하지 않음. */
	options.EncodingType = EncodingTypeNone

	t0 := time.Now()
	for {
		tl0 := time.Now()
		content, err := RequestHttp(url, options)
		if err == nil {
			if len(content) == 0 {
				return nil, errors.New("no peer found")
			}

			var peers []string
			for _, peerToken := range strings.Split(string(content), ",") {
				peerHost := strings.TrimSpace(peerToken)
				if peerHost == "" {
					continue
				}
				peers = append(peers, peerHost)
			}
			return peers, nil
		}
		log.Error().Err(err).Msg("An error occured during get peer list.")

		tl1 := time.Now()
		if tl1.Sub(t0) > timeout {
			break
		}

		elapsed := tl1.Sub(tl0)
		if elapsed < InitScanPeerInterval {
			time.Sleep(InitScanPeerInterval - elapsed)
		}
	}

	return nil, errors.New("timeout waiting for peer list")
}

func (mainContext *MainContext) Init() (err error) {
	defer func() {
		if err != nil {
			log.Error().Err(err).Msg("Initialize failed")
		}
	}()

	mainContext.Hostname, err = os.Hostname()
	if err != nil {
		return
	}

	mainContext.UUID = uuid.New().String()

	mainContext.IdleTimeout = *flagTransferIdleTimeout
	mainContext.PrintStatistics = *flagStatistics
	if mainContext.PrintStatistics {
		mainContext.Statistics = make(map[string]string)
	}

	if !IsAvailableEncoding(flagCompressType) {
		err = fmt.Errorf("unsupported compress type.:%v", *flagCompressType)
		return
	}

	normalizeDir := func(dir string) string {
		if dir == "" {
			return ""
		}

		normalized, err := filepath.Abs(dir)
		if err != nil {
			dir = strings.TrimSpace(dir)
			normalized = strings.TrimSuffix(dir, "/")
		}
		return normalized
	}

	if *flagChunkSize <= 0 || *flagChunkSize > MaxChunkSize {
		err = fmt.Errorf("invalid chunk size parameter.:%v", *flagChunkSize)
		return
	}

	mainContext.DirSrc = normalizeDir(*flagSrc)
	mainContext.DirDst = normalizeDir(*flagDst)

	if mainContext.DirSrc != "" && mainContext.DirDst == "" {
		// src 만 있어도 배포는 가능하나
		// src 는 보통 network drive 이고 너무 많은 접근으로 인해 peer 전송보다 성능이 떨어지게됨
		// dst 에 복사된 파일을 배포해야 성능이 더 좋음
		mainContext.Completed.Store(true)
		mainContext.ServeOnly = true
	}

	mainContext.PeerList = make([]string, 0)
	peerTokens := strings.Split(*flagPeerList, ",")
	for _, peerToken := range peerTokens {
		peerHost := strings.TrimSpace(peerToken)
		if peerHost == "" {
			continue
		}
		mainContext.PeerList = append(mainContext.PeerList, peerHost)
	}

	if *flagPeerListURL != "" {
		var list []string
		list, err = peerListFromURL(mainContext, strings.TrimSpace(*flagPeerListURL), *flagPeerWaitTimeout)
		if err != nil {
			log.Error().Err(err).Str("URL", *flagPeerListURL).Msg("Failed to retrieve peers list from URL")
			return
		}
		mainContext.PeerList = append(mainContext.PeerList, list...)
	}

	mainContext.TransferChunks, err = mainContext.Manifest.Init(mainContext.DirSrc, mainContext.PeerList)
	if err != nil {
		return
	}

	if *flagPeerNumConcurrent < 1 || *flagPeerNumConcurrent > MaxPeerNumConcurrent {
		err = fmt.Errorf("invalid concurrent download jobs parameter.:%v", *flagPeerNumConcurrent)
		return
	}
	mainContext.PeerNumConcurrent = *flagPeerNumConcurrent

	if *flagLocalNumConcurrent < 1 || *flagLocalNumConcurrent > MaxLocalNumConcurrent {
		err = fmt.Errorf("invalid concurrent download jobs parameter.:%v", *flagLocalNumConcurrent)
		return
	}
	mainContext.LocalNumConcurrent = *flagLocalNumConcurrent

	log.Info().Msg("Initialization done.")
	return
}

func createDir(mainContext *MainContext, dir types.Dir) error {
	pathToCreated := mainContext.DirDst + "/" + dir.Name

	/* 이미 생성된 경로에 권한 부여 */
	os.Chmod(pathToCreated, 0700)
	return os.MkdirAll(pathToCreated, 0700)
}

func createSymlink(mainContext *MainContext, path string, symlinkTo string) error {
	pathToCreated := mainContext.DirDst + "/" + path

	/* 파일이 이미 존재하는 경우 삭제 후 생성 */
	_, err := os.Lstat(pathToCreated)
	if err == nil {
		err = os.Remove(pathToCreated)
		if err != nil {
			return err
		}
	} else {
		if os.IsNotExist(err) {
			err = nil
		} else {
			return err
		}
	}

	return os.Symlink(symlinkTo, pathToCreated)
}

func createEmptyFile(mainContext *MainContext, emptyFile types.EmptyFile) error {
	pathToCreated := mainContext.DirDst + "/" + emptyFile.Name

	/* 이미 생성된 파일을 truncate 하기 위한 권한 부여 */
	os.Chmod(pathToCreated, 0600)
	localFile, err := os.OpenFile(pathToCreated, os.O_CREATE|os.O_TRUNC, emptyFile.Perm)
	if err != nil {
		return err
	}

	return localFile.Close()
}

func createFile(mainContext *MainContext, file types.File) error {
	pathToCreated := mainContext.DirDst + "/" + file.Name

	/* 이미 생성된 파일에 read, write 를 위한 권한 부여 */
	os.Chmod(pathToCreated, 0600)
	localFile, err := os.OpenFile(pathToCreated, os.O_CREATE|os.O_RDWR, 0600)
	if err != nil {
		return err
	}
	defer localFile.Close()

	return syscall.Ftruncate(int(localFile.Fd()), file.Size)
}

type ChunkQueue struct {
	q     []int
	mutex sync.RWMutex
}

func (queue *ChunkQueue) Len() int {
	queue.mutex.RLock()
	defer queue.mutex.RUnlock()
	return len(queue.q)
}

func (queue *ChunkQueue) PushBack(transferChunkIndex int) {
	queue.mutex.Lock()
	defer queue.mutex.Unlock()
	queue.q = append(queue.q, transferChunkIndex)
}

func (queue *ChunkQueue) PushShuffle(transferChunkIndex int) {
	randomVal := rand.Int()

	queue.mutex.Lock()
	defer queue.mutex.Unlock()

	queue.q = append(queue.q, transferChunkIndex)

	lenQueue := len(queue.q)
	indexInsert := randomVal % (lenQueue)
	if indexInsert == lenQueue {
		return
	}
	(queue.q)[indexInsert], (queue.q)[lenQueue-1] = (queue.q)[lenQueue-1], (queue.q)[indexInsert]
}

func (queue *ChunkQueue) Pop() (int, bool) {
	queue.mutex.Lock()
	defer queue.mutex.Unlock()

	if len(queue.q) != 0 {
		transferChunkIndex := (queue.q)[0]
		queue.q = (queue.q)[1:]
		return transferChunkIndex, true
	}

	return 0, false
}

func writeStreamToFile(mainContext *MainContext, transferChunkIndex int, reader io.Reader) (int64, error) {
	var length int64
	for _, fileChunk := range mainContext.TransferChunks[transferChunkIndex].FileChunks {
		file := &mainContext.Manifest.Files[fileChunk.FileIndex]
		n, err := func() (n int64, err error) {
			f, err := os.OpenFile(mainContext.DirDst+file.Name, os.O_WRONLY, 0)
			if err != nil {
				return
			}
			defer f.Close()

			_, err = f.Seek(fileChunk.FromOffset, io.SeekStart)
			if err != nil {
				return
			}

			return io.CopyN(f, reader, fileChunk.ToOffset-fileChunk.FromOffset)
		}()

		if err != nil {
			return 0, err
		}

		length += n
	}

	return length, nil
}

func execDownloader(mainContext *MainContext, downloader FileDownloader, numPeerConcurrent int) {
	defer mainContext.WGDownloader.Done()
	defer downloader.Fin()

	/* 통계정보 - critical section */
	var bytesReceivedTotal atomic.Int64
	var chunksReceivedTotal atomic.Int64
	var durationTotal utilx.AtomicDuration
	defer func() {
		if !mainContext.PrintStatistics {
			return
		}
		mainContext.MutexStatistics.Lock()
		mainContext.Statistics[fmt.Sprintf("[%v] received_bytes", downloader.Name())] = fmt.Sprint(bytesReceivedTotal.Load())
		mainContext.Statistics[fmt.Sprintf("[%v] received_chunks", downloader.Name())] = fmt.Sprint(chunksReceivedTotal.Load())
		mainContext.Statistics[fmt.Sprintf("[%v] cumulative_transfer_time", downloader.Name())] = durationTotal.Load().String()
		mainContext.BytesReceivedTotal.Add(bytesReceivedTotal.Load())
		mainContext.ChunksReceivedTotal.Add(chunksReceivedTotal.Load())
		mainContext.CumulativeDurationTotal.Add(durationTotal.Load())
		mainContext.MutexStatistics.Unlock()
	}()

	var queue ChunkQueue
	var numCompleted atomic.Int64

	var wgWorker sync.WaitGroup
	defer wgWorker.Wait()

	// 전송해야 하는 chunk 를 downloadWorker 고루틴에 전달하기 위한 채널 생성
	chanRequest := make(chan int)
	defer close(chanRequest)

	nameDownloader := downloader.Name()

	downloadWorker := func() {
		defer wgWorker.Done()
		for {
			transferChunkIndex, ok := <-chanRequest
			if !ok {
				break
			}

			t0 := time.Now()
			length, err := func() (int64, error) {
				reader, err := downloader.RequestDownload(transferChunkIndex)
				if err != nil {
					return 0, err
				}
				defer reader.Close()

				return writeStreamToFile(mainContext, transferChunkIndex, reader)
			}()
			t1 := time.Now()

			mainContext.LastUpdatedAt.Store(t1)

			if err == nil {
				log.Info().Str("host", mainContext.Hostname).Str("downloader", nameDownloader).Int("chunk", transferChunkIndex).Str("elapsed", t1.Sub(t0).String()).Msg("downloaded")
				chunksReceivedTotal.Add(1)
				bytesReceivedTotal.Add(length)
				durationTotal.Add(t1.Sub(t0))

				mainContext.TransferChunks[transferChunkIndex].LastUpdatedAt.Store(t1)
				mainContext.TransferChunks[transferChunkIndex].Status.Store(types.ChunkStatusDone)

				numCompleted.Add(1)
			} else {
				log.Error().Str("host", mainContext.Hostname).Str("downloader", nameDownloader).Int("chunk", transferChunkIndex).Err(err).Msg("download failed")

				/* 결과가 error인 경우 재시도 */
				mainContext.TransferChunks[transferChunkIndex].Status.Store(types.ChunkStatusPending)
				queue.PushBack(transferChunkIndex)
			}
		}
	}
	wgWorker.Add(numPeerConcurrent)
	for i := 0; i < numPeerConcurrent; i++ {
		go downloadWorker()
	}

	for {
		if !downloader.IsValid() {
			break
		}

		if mainContext.WantStopDownload.Load() {
			break
		}

		/* 이 Downloader가 완료한 경우 */
		if mainContext.Manifest.TransferChunkCount == int(numCompleted.Load()) {
			mainContext.WantStopDownload.Store(true)
			mainContext.Completed.Store(true)
			log.Info().Str("host", mainContext.Hostname).Str("downloader", nameDownloader).Msg("download job done.")
			break
		}

		if len(chanRequest) != 0 {
			time.Sleep(1 * time.Millisecond)
			continue
		}

		/* random 순서로 queue */
		downloader.FlushAvailableList(&queue)

		transferChunkIndex, ok := queue.Pop()
		if ok {
			/* Pending 상태인 경우 상태 변경하여 표시하고 download 시도 */
			switch mainContext.TransferChunks[transferChunkIndex].Status.Load() {
			case types.ChunkStatusPending:
				if mainContext.TransferChunks[transferChunkIndex].Status.CompareAndSwap(types.ChunkStatusPending, types.ChunkStatusDownloading) {
					/* 전송해야 하는 Chunk. */
					chanRequest <- transferChunkIndex
				} else {
					/* 다른 downloader에 의해 상태가 변경됨. 재시도 */
					queue.PushBack(transferChunkIndex)
				}

			case types.ChunkStatusDownloading:
				/* 다른 downloader에 의해서 전송중인 Chunk. 이 경우 보류 */
				queue.PushBack(transferChunkIndex)

			case types.ChunkStatusDone:
				/* 다른 downloader에 의해 이미 전송 완료된 Chunk. 이 downloader에서도 완료한 것으로 함. */
				numCompleted.Add(1)
			}
		} else {
			/* queue에 새로운 것이 없다면 sleep */
			time.Sleep(DurationDownloaderSleep)
		}
	}
}

func DownloadFiles(mainContext *MainContext) error {
	/* 링크와 파일 다운로드 준비를 위해 디렉토리 먼저 생성 */
	err := os.MkdirAll(mainContext.DirDst, os.ModePerm)
	if err != nil {
		return err
	}
	for _, dir := range mainContext.Manifest.Dirs {
		err = createDir(mainContext, dir)
		if err != nil {
			log.Error().Err(err).Str("dir", dir.Name).Msg("create directory failed")
			return err
		}
		log.Info().Str("dir", dir.Name).Msg("created directory")
	}

	for _, symlink := range mainContext.Manifest.Symlinks {
		err = createSymlink(mainContext, symlink.Name, symlink.SymlinkTo)
		if err != nil {
			log.Error().Err(err).Str("symlink", symlink.Name).Msg("create symbolic link failed")
			return err
		}
		log.Info().Str("symlink", symlink.Name).Msg("created symbolic link")
	}

	/* 파일 길이가 0인 파일은 다운로드 하지 않고 그냥 생성 */
	for _, emptyFile := range mainContext.Manifest.EmptyFiles {
		err = createEmptyFile(mainContext, emptyFile)
		if err != nil {
			log.Error().Err(err).Str("file", emptyFile.Name).Msg("create empty file failed")
			return err
		}
		log.Info().Str("file", emptyFile.Name).Msg("created empty file")
	}

	/* 전송할 파일을 미리 만들고 resize 함 */
	for _, file := range mainContext.Manifest.Files {
		err = createFile(mainContext, file)
		if err != nil {
			log.Error().Err(err).Str("file", file.Name).Msg("create file failed")
			return err
		}
	}

	t0 := time.Now()
	/* 각 worker에서 download 성공할 때 마다 이 변수를 해당 시점 시간으로 변경 */
	mainContext.LastUpdatedAt.Store(t0)
	for _, peerHost := range mainContext.PeerList {
		var peerDownloader PeerFileDownloader
		peerDownloader.Init(mainContext, peerHost)
		mainContext.WGDownloader.Add(1)
		go execDownloader(mainContext, &peerDownloader, mainContext.PeerNumConcurrent)
	}

	if mainContext.DirSrc != "" {
		var localDownloader LocalFileDownloader
		localDownloader.Init(mainContext)
		mainContext.WGDownloader.Add(1)
		go execDownloader(mainContext, &localDownloader, mainContext.LocalNumConcurrent)
	}

	/* 대기상태 timeout 모니터. timeout인 경우 err을 이 함수에서 설정. */
	if mainContext.IdleTimeout > 0 {
		mainContext.WGDownloader.Add(1)
		go func() {
			defer mainContext.WGDownloader.Done()
			ticker := time.NewTicker(DurationDownloaderSleep)
			defer ticker.Stop()
			for ; ; <-ticker.C {
				if mainContext.WantStopDownload.Load() {
					return
				}

				if time.Since(mainContext.LastUpdatedAt.Load()) >= mainContext.IdleTimeout {
					mainContext.WantStopDownload.Store(true)
					err = fmt.Errorf("transfer idle timeout %v exceeded", mainContext.IdleTimeout)
					return
				}
			}
		}()
	}
	mainContext.WGDownloader.Wait()

	// 전송이 끝난 후 권한 부여
	for _, file := range mainContext.Manifest.Files {
		os.Chmod(mainContext.DirDst+"/"+file.Name, file.Perm)
	}

	for _, dir := range mainContext.Manifest.Dirs {
		os.Chmod(mainContext.DirDst+"/"+dir.Name, dir.Perm)
	}

	t1 := time.Now()
	elapsed := t1.Sub(t0)

	if mainContext.PrintStatistics {
		mainContext.Statistics["[Total] received_bytes"] = fmt.Sprint(mainContext.BytesReceivedTotal.Load())
		mainContext.Statistics["[Total] received_chunks"] = fmt.Sprint(mainContext.ChunksReceivedTotal.Load())
		mainContext.Statistics["[Total] cumulative_transfer_time"] = mainContext.CumulativeDurationTotal.Load().String()
		mainContext.Statistics["[Total] elapsed_time"] = elapsed.String()

		elapsedSeconds := int64(elapsed.Seconds())
		if elapsedSeconds == 0 {
			elapsedSeconds = 1
		}
		mainContext.Statistics["[Total] transfer_bytes_per_seconds"] = fmt.Sprint(mainContext.BytesReceivedTotal.Load() / elapsedSeconds)
		printStatistics(mainContext)
	}

	log.Info().Str("host", mainContext.Hostname).Str("elapsed", elapsed.String()).Msg("download job done.")
	return nil
}

func waitHost(host string, timeout time.Duration) error {
	/* /completed 가 200을 반환할 때 까지 시도 */
	t0 := time.Now()
	ticker := time.NewTicker(DurationDownloaderSleep)
	defer ticker.Stop()
	for ; time.Since(t0) < timeout; <-ticker.C {
		url := fmt.Sprintf("http://%v/completed", host)
		options := NewHttpRequestOptions()
		options.ExpectedContentType = "text/plain"
		_, err := RequestHttp(url, options)
		if err == nil {
			return nil
		}
	}

	return fmt.Errorf("timeout for the waiting host %v exceeded", host)
}

func getVersion() string {
	return fmt.Sprintf("p2pcp-%s.%s.%s+%s", MajorVersion, MinorVersion, PatchVersion, BuildVersion)
}

func printStatistics(mainContext *MainContext) {
	var keys []string
	for key := range mainContext.Statistics {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	fmt.Printf("%s download staistics information:\n", getVersion())
	for _, key := range keys {
		value := mainContext.Statistics[key]
		fmt.Printf("[%v] %v = %v\n", mainContext.Hostname, key, value)
	}
}
