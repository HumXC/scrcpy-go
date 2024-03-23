package option

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
)

// Base: https://github.com/Genymobile/scrcpy/blob/1ee46970e373ea3c34c3d9b632fef34982d7a52b/server/src/main/java/com/genymobile/scrcpy/Options.java
type Options struct {
	Scid                  int    `scrcpy_opt:"scid"`
	LogLevel              string `scrcpy_opt:"log_level"`
	Video                 bool   `scrcpy_opt:"video"`
	Audio                 bool   `scrcpy_opt:"audio"`
	VideoCodec            string `scrcpy_opt:"video_codec"`
	AudioCodec            string `scrcpy_opt:"audio_codec"`
	AudioSource           string `scrcpy_opt:"audio_source"`
	MaxSize               int    `scrcpy_opt:"max_size"`
	VideoBitRate          int    `scrcpy_opt:"video_bit_rate"`
	AudioBitRate          int    `scrcpy_opt:"audio_bit_rate"`
	MaxFps                int    `scrcpy_opt:"max_fps"`
	LockVideoOrientation  int    `scrcpy_opt:"lock_video_orientation"`
	TunnelForward         bool   `scrcpy_opt:"tunnel_forward"`
	Crop                  string `scrcpy_opt:"crop"`
	Control               bool   `scrcpy_opt:"control"`
	DisplayId             int    `scrcpy_opt:"display_id"`
	ShowTouches           bool   `scrcpy_opt:"show_touches"`
	StayAwake             bool   `scrcpy_opt:"stay_awake"`
	VideoCodecOptions     string `scrcpy_opt:"video_codec_options"`
	AudioCodecOptions     string `scrcpy_opt:"audio_codec_options"`
	VideoEncoder          string `scrcpy_opt:"video_encoder"`
	AudioEncoder          string `scrcpy_opt:"audio_encoder"`
	PowerOffScreenOnClose bool   `scrcpy_opt:"power_off_on_close"`
	ClipboardAutosync     bool   `scrcpy_opt:"clipboard_autosync"`
	DownsizeOnError       bool   `scrcpy_opt:"downsize_on_error"`
	Cleanup               bool   `scrcpy_opt:"cleanup"`
	PowerOn               bool   `scrcpy_opt:"power_on"`
	ListEncoders          bool   `scrcpy_opt:"list_encoders"`
	ListDisplays          bool   `scrcpy_opt:"list_displays"`
	SendDeviceMeta        bool   `scrcpy_opt:"send_device_meta"`
	SendFrameMeta         bool   `scrcpy_opt:"send_frame_meta"`
	SendDummyByte         bool   `scrcpy_opt:"send_dummy_byte"`
	SendCodecMeta         bool   `scrcpy_opt:"send_codec_meta"`
	RawStream             bool   `scrcpy_opt:"raw_stream"`
}

// 返回用于启动 scrcpy-server 的参数
// 当字段的值与默认值相同时会被忽略
// 忽略 TunnelForward, TunnelForward 必须为 true
func (s *Options) ToArgs() []string {
	// tunnel_forward 必须开启
	o := *s
	o.TunnelForward = true
	default_ := Default()
	defaultV := reflect.ValueOf(default_)
	optV := reflect.ValueOf(&o).Elem()
	optT := reflect.TypeOf(o)
	args := make([]string, 0)
	for i := 0; i < optV.NumField(); i++ {
		fie := optV.Field(i)
		// 遍历字段，如果 optV 与 defaultV 里字段的值相同，则跳过
		if fie.Interface() == defaultV.Field(i).Interface() {
			continue
		}
		var value any
		switch fie.Kind() {
		case reflect.Bool:
			value = fie.Bool()
		case reflect.String:
			value = fie.String()
		case reflect.Int:
			value = fie.Int()
		default:
			panic("unsupported type: " + fie.Kind().String())
		}
		args = append(args, fmt.Sprintf("%s=%v", optT.Field(i).Tag.Get("scrcpy_opt"), value))
	}
	return args
}

// Default() 返回一个带有默认值的 ScrcpyOptions
// 其默认值不是 go 的默认值，而是 scrcpy-server 中定义的默认值，与 scrcpy-server 中的默认值保持一致
// https://github.com/Genymobile/scrcpy/blob/1ee46970e373ea3c34c3d9b632fef34982d7a52b/server/src/main/java/com/genymobile/scrcpy/Options.java#L8
func Default() Options {
	return Options{
		// TODO：
		// private VideoCodec videoCodec = VideoCodec.H264;
		// private AudioCodec audioCodec = AudioCodec.OPUS;
		// private AudioSource audioSource = AudioSource.OUTPUT;
		Scid:                 -1, // 31-bit non-negative value, or -1
		Audio:                true,
		Video:                true,
		VideoBitRate:         8000000,
		AudioBitRate:         128000,
		LockVideoOrientation: -1,
		// TODO:
		// private List<CodecOption> videoCodecOptions;
		// private List<CodecOption> audioCodecOptions;
		Control:           true,
		ClipboardAutosync: true,
		DownsizeOnError:   true,
		Cleanup:           true,
		PowerOn:           true,
		SendDeviceMeta:    true,
		SendFrameMeta:     true,
		SendDummyByte:     true,
		SendCodecMeta:     true,
	}
}

// 忽略 TunnelForward, TunnelForward 必须为 true
func Parse(args []string) (Options, error) {
	opt := Default()
	optV := reflect.ValueOf(&opt).Elem()
	optT := reflect.TypeOf(opt)
	keyValue := make(map[string]reflect.Value)
	for i := 0; i < optV.NumField(); i++ {
		field := optV.Field(i)
		tag := optT.Field(i).Tag.Get("scrcpy_opt")
		if tag == "" {
			panic("has field without \"scrcpy_opt\" tag in options.ScrcpyOptions." + optT.Field(i).Name)
		}
		keyValue[tag] = field
	}

	for _, arg := range args {
		if arg == "" {
			continue
		}
		kv := strings.Split(arg, "=")
		if len(kv) < 2 {
			continue
		}
		if v, ok := keyValue[kv[0]]; ok {
			err := setValue(v, kv[1])
			if err != nil {
				return opt, fmt.Errorf("value conver error: %w", err)
			}
		} else {
			return opt, fmt.Errorf("unknown option: %s", arg)
		}
	}
	finalOpt := optV.Interface().(Options)
	finalOpt.TunnelForward = true
	return finalOpt, nil
}
func setValue(field reflect.Value, value string) error {
	switch field.Kind() {
	case reflect.Bool:
		v, err := strconv.ParseBool(value)
		if err != nil {
			return err
		}
		field.SetBool(v)
	case reflect.Int:
		v, err := strconv.ParseInt(value, 10, 64)
		if err != nil {
			return err
		}
		field.SetInt(v)
	case reflect.String:
		field.SetString(value)
	default:
		// 能走到这说明 ScrcpyOptions 结构体有问题
		panic("unsupported type: " + field.Kind().String())
	}
	return nil
}
