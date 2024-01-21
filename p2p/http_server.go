package p2p

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/gokch/memechain/types"
	"github.com/rs/zerolog/log"
)

func textResponse(w http.ResponseWriter, content string, status int) error {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.Header().Set("Content-Length", fmt.Sprintf("%d", len(content)))
	w.WriteHeader(status)

	written, err := w.Write([]byte(content))
	mainContext.BytesTransmittedTotal.Add(int64(written))

	return err
}

func errorResponse(w http.ResponseWriter, status int, err error) {
	textResponse(w, err.Error(), status)
}

func setEncodingType(w http.ResponseWriter, r *http.Request) string {
	encodingType := GetPreferredEncodingType(r.Header["Accept-Encoding"])
	if encodingType != EncodingTypeNone {
		w.Header().Set("Content-Encoding", encodingType)
	}
	return encodingType
}

func CloseServer(mainContext *MainContext, server *http.Server) error {
	shutdownContext, cancel := context.WithTimeout(context.Background(), ShutdownTimeout)
	defer cancel()
	return server.Shutdown(shutdownContext)
}

func CreateHttpServer(mainContext *MainContext) *http.Server {

	mux := http.NewServeMux()

	/* 진행사항 모니터링 위한 핸들러 */
	mux.HandleFunc("/completed", func(w http.ResponseWriter, r *http.Request) {
		var status int
		var content string
		if !mainContext.Completed.Load() {
			status = http.StatusNotFound
			content = "Downloading"
		} else {
			status = http.StatusOK
			content = "OK"
		}

		err := textResponse(w, content, status)
		if err != nil {
			log.Error().Err(err).Msg("Write failed.")
			return
		}
	})

	/* 연결된 peer가 자기자신인지 확인하기 위한 핸들러 */
	mux.HandleFunc("/uuid", func(w http.ResponseWriter, r *http.Request) {
		err := textResponse(w, mainContext.UUID, http.StatusOK)
		if err != nil {
			log.Error().Err(err).Msg("Write failed.")
			return
		}
	})

	mux.HandleFunc("/version", func(w http.ResponseWriter, r *http.Request) {
		err := textResponse(w, getVersion(), http.StatusOK)
		if err != nil {
			log.Error().Err(err).Msg("Write failed.")
			return
		}
	})

	mux.HandleFunc("/manifest", func(w http.ResponseWriter, r *http.Request) {
		encodingType := setEncodingType(w, r)
		reader, length, err := NewBufferEncodingReader(mainContext.Manifest.JsonData, encodingType)
		if err != nil {
			log.Error().Err(err).Msg("NewBufferEncodingReader failed.")
			errorResponse(w, http.StatusInternalServerError, err)
			return
		}
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.Header().Set("Content-Length", fmt.Sprintf("%d", length))

		written, err := io.Copy(w, reader)
		mainContext.BytesTransmittedTotal.Add(written)
		if err != nil {
			log.Error().Err(err).Msg("io.Copy failed.")
			errorResponse(w, http.StatusInternalServerError, err)
			return
		}
	})

	mux.HandleFunc("/manifest/checksum", func(w http.ResponseWriter, r *http.Request) {
		err := textResponse(w, mainContext.Manifest.ChunkChecksum, http.StatusOK)
		if err != nil {
			log.Error().Err(err).Msg("Write failed.")
			return
		}
	})

	mux.HandleFunc("/chunk", func(w http.ResponseWriter, r *http.Request) {
		onlyUpdatedSinceArray := r.URL.Query()["only_updated_since"]
		var onlyUpdatedSinceStr string
		if len(onlyUpdatedSinceArray) > 0 {
			onlyUpdatedSinceStr = onlyUpdatedSinceArray[0]
		}

		now := time.Now().Format(time.RFC3339Nano)
		resultObject := types.AvailableChunkList{
			Timestamp: now,
		}

		if mainContext.Completed.Load() {
			resultObject.IsCompleted = true
		} else {
			var onlyUpdatedSince time.Time
			if onlyUpdatedSinceStr != "" {
				var err error
				onlyUpdatedSince, err = time.Parse(time.RFC3339Nano, onlyUpdatedSinceStr)
				if err != nil {
					log.Error().Err(err).Msgf("invalid timestamp.:only_updated_since=%v", onlyUpdatedSinceStr)
					errorResponse(w, http.StatusBadRequest, err)
					return
				}
			}

			for index := range mainContext.TransferChunks {
				if mainContext.TransferChunks[index].Status.Load() == types.ChunkStatusDone &&
					onlyUpdatedSince.After(mainContext.TransferChunks[index].LastUpdatedAt.Load()) {
					resultObject.TransferChunkIndexList = append(resultObject.TransferChunkIndexList, index)
				}
			}
		}

		json, err := json.Marshal(resultObject)
		if err != nil {
			log.Error().Err(err).Msg("JSON Marshal failed.")
			errorResponse(w, http.StatusInternalServerError, err)
			return
		}

		encodingType := setEncodingType(w, r)
		reader, length, err := NewBufferEncodingReader(json, encodingType)
		if err != nil {
			log.Error().Err(err).Msg("NewBufferEncodingReader failed.")
			errorResponse(w, http.StatusInternalServerError, err)
			return
		}
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.Header().Set("Content-Length", fmt.Sprintf("%d", length))

		written, err := io.Copy(w, reader)
		mainContext.BytesTransmittedTotal.Add(written)
		if err != nil {
			log.Error().Err(err).Msg("io.Copy failed.")
			errorResponse(w, http.StatusInternalServerError, err)
			return
		}
	})

	const servePointFiles = "/chunk/"
	var htdocs string
	var GetChunkStatus func(int) int32
	if mainContext.ServeOnly {
		htdocs = mainContext.DirSrc
		GetChunkStatus = func(int) int32 { return types.ChunkStatusDone }
	} else {
		htdocs = mainContext.DirDst
		GetChunkStatus = func(transferChunkIndex int) int32 {
			return mainContext.TransferChunks[transferChunkIndex].Status.Load()
		}
	}
	mux.HandleFunc(servePointFiles, func(w http.ResponseWriter, r *http.Request) {
		path := strings.TrimPrefix(r.URL.Path, servePointFiles)
		transferChunkIndex, err := strconv.Atoi(path)
		if err != nil {
			log.Error().Err(err).Msgf("strconv.Atoi failed. got (%s)", r.URL.Path)
			errorResponse(w, http.StatusForbidden, errors.New("chunk index is not an integer"))
			return
		}

		if transferChunkIndex < 0 || transferChunkIndex >= mainContext.Manifest.TransferChunkCount {
			log.Error().Msgf("chunk index is out of range. should be in range 0 ~ %d. got (%d)", mainContext.Manifest.TransferChunkCount, transferChunkIndex)
			errorResponse(w, http.StatusForbidden, errors.New("chunk index out of range"))
			return
		}

		status := GetChunkStatus(transferChunkIndex)
		if status != types.ChunkStatusDone {
			log.Error().Msgf("chunk not ready: chunk:%v, %v", transferChunkIndex, status)
			errorResponse(w, http.StatusForbidden, errors.New("chunk not ready"))
			return
		}

		encodingType := setEncodingType(w, r)
		writer := NewEncodingWriteCloser(w, encodingType)
		defer writer.Close()

		w.Header().Set("Content-Type", "application/octet-stream")
		if encodingType == EncodingTypeNone {
			length := mainContext.TransferChunks[transferChunkIndex].Size()
			w.Header().Set("Content-Length", fmt.Sprintf("%d", length))
		}

		for _, fileChunk := range mainContext.TransferChunks[transferChunkIndex].FileChunks {
			file := &mainContext.Manifest.Files[fileChunk.FileIndex]

			f, err := os.Open(htdocs + file.Name)
			if err != nil {
				log.Error().Err(err).Str("fileName", htdocs+file.Name).Msg("Open failed.")
				errorResponse(w, http.StatusInternalServerError, err)
				return
			}
			defer f.Close()

			_, err = f.Seek(fileChunk.FromOffset, io.SeekStart)
			if err != nil {
				log.Error().Err(err).Str("fileName", htdocs+file.Name).Msg("Seek failed.")
				errorResponse(w, http.StatusInternalServerError, err)
				return
			}

			written, err := io.CopyN(writer, f, fileChunk.ToOffset-fileChunk.FromOffset)
			if err != nil {
				log.Error().Err(err).Str("fileName", htdocs+file.Name).Msg("CopyN failed.")
				errorResponse(w, http.StatusInternalServerError, err)
				return
			}
			mainContext.BytesTransmittedTotal.Add(written)
		}
	})

	return &http.Server{
		Addr:         *flagListenAddr,
		Handler:      mux,
		ReadTimeout:  1 * time.Second,
		WriteTimeout: 60 * time.Second,
	}
}
