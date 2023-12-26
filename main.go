package record

import (
	_ "embed"
	"errors"
	"io"
	"sync"

	"go.uber.org/zap"
	. "m7s.live/engine/v4"
	"m7s.live/engine/v4/codec"
	"m7s.live/engine/v4/config"
	"m7s.live/engine/v4/util"
)

type RecordConfig struct {
	DefaultYaml
	config.Subscribe
	Flv        Record
	Mp4        Record
	Fmp4       Record
	Hls        Record
	Raw        Record
	RawAudio   Record
	recordings sync.Map
}

//go:embed default.yaml
var defaultYaml DefaultYaml
var ErrRecordExist = errors.New("recorder exist")
var RecordPluginConfig = &RecordConfig{
	DefaultYaml: defaultYaml,
	Flv: Record{
		Path:          "record/flv",
		Ext:           ".flv",
		GetDurationFn: getFLVDuration,
	},
	Fmp4: Record{
		Path: "record/fmp4",
		Ext:  ".mp4",
	},
	Mp4: Record{
		Path: "record/mp4",
		Ext:  ".mp4",
	},
	Hls: Record{
		Path: "record/hls",
		Ext:  ".m3u8",
	},
	Raw: Record{
		Path: "record/raw",
		Ext:  ".", // 默认h264扩展名为.h264,h265扩展名为.h265
	},
	RawAudio: Record{
		Path: "record/raw",
		Ext:  ".", // 默认aac扩展名为.aac,pcma扩展名为.pcma,pcmu扩展名为.pcmu
	},
}

var plugin = InstallPlugin(RecordPluginConfig)

func (conf *RecordConfig) OnEvent(event any) {
	// switch v := event.(type) {
	switch event.(type) {
	case FirstConfig, config.Config:
		conf.Flv.Init()
		conf.Mp4.Init()
		conf.Fmp4.Init()
		conf.Hls.Init()
		conf.Raw.Init()
		conf.RawAudio.Init()
	case SEpublish:
		// streamPath := v.Target.Path
		// if conf.Flv.NeedRecord(streamPath) {
		// 	go NewFLVRecorder().Start(streamPath)
		// }
		// if conf.Mp4.NeedRecord(streamPath) {
		// 	go NewMP4Recorder().Start(streamPath)
		// }
		// if conf.Fmp4.NeedRecord(streamPath) {
		// 	go NewFMP4Recorder().Start(streamPath)
		// }
		// if conf.Hls.NeedRecord(streamPath) {
		// 	go NewHLSRecorder().Start(streamPath)
		// }
		// if conf.Raw.NeedRecord(streamPath) {
		// 	go NewRawRecorder().Start(streamPath)
		// }
		// if conf.RawAudio.NeedRecord(streamPath) {
		// 	go NewRawAudioRecorder().Start(streamPath)
		// }
	}
}
func (conf *RecordConfig) getRecorderConfigByType(t string) (recorder *Record) {
	switch t {
	case "flv":
		recorder = &conf.Flv
	case "mp4":
		recorder = &conf.Mp4
	case "fmp4":
		recorder = &conf.Fmp4
	case "hls":
		recorder = &conf.Hls
	case "raw":
		recorder = &conf.Raw
	case "raw_audio":
		recorder = &conf.RawAudio
	}
	return
}

func (conf *RecordConfig) Start(streamPath string, format string) (string, error) {
	var id string
	var err error

	var irecorder IRecorder
	switch format {
	case "flv":
		irecorder = NewFLVRecorder()
	case "mp4":
		irecorder = NewMP4Recorder()
	case "fmp4":
		irecorder = NewFMP4Recorder()
	case "hls":
		irecorder = NewHLSRecorder()
	case "raw":
		irecorder = NewRawRecorder()
	case "raw_audio":
		irecorder = NewRawAudioRecorder()
	default:
		err = errors.New("type not supported")
	}
	if err != nil {
		return id, err
	}

	id = irecorder.GetRecorder().ID
	err = irecorder.Start(streamPath)

	return id, err
}

func (conf *RecordConfig) Stop(id string) {
	if recorder, ok := conf.recordings.LoadAndDelete(id); ok {
		recorder.(ISubscriber).Stop(zap.String("reason", "api stop"))
	}
}

func getFLVDuration(file io.ReadSeeker) uint32 {
	_, err := file.Seek(-4, io.SeekEnd)
	if err == nil {
		var tagSize uint32
		if tagSize, err = util.ReadByteToUint32(file, true); err == nil {
			_, err = file.Seek(-int64(tagSize)-4, io.SeekEnd)
			if err == nil {
				_, timestamp, _, err := codec.ReadFLVTag(file)
				if err == nil {
					return timestamp
				}
			}
		}
	}
	return 0
}
