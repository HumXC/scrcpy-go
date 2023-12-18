package codecs

type Codec struct {
	Id   uint32
	Name string
}

func (c Codec) String() string {
	return c.Name
}

// https://github.com/Genymobile/scrcpy/blob/a3cdf1a6b86ea22786e1f7d09b9c202feabc6949/server/src/main/java/com/genymobile/scrcpy/AudioCodec.java
// https://github.com/Genymobile/scrcpy/blob/a3cdf1a6b86ea22786e1f7d09b9c202feabc6949/server/src/main/java/com/genymobile/scrcpy/AudioCodec.java
var (
	VideoH264 = Codec{0x68_32_36_34, "h264"}
	VideoH265 = Codec{0x68_32_36_35, "h265"}
	VideoAV1  = Codec{0x00_61_76_31, "av1"}

	AudioAAC  = Codec{0x00_61_61_63, "aac"}
	AudioOPUS = Codec{0x6f_70_75_73, "opus"}
	AudioRAW  = Codec{0x00_72_61_77, "raw"}
)

var allcodec = []Codec{
	VideoH264,
	VideoH265,
	VideoAV1,
	AudioAAC,
	AudioOPUS,
	AudioRAW,
}

func FromId(id uint32) Codec {
	for _, codec := range allcodec {
		if codec.Id == id {
			return codec
		}
	}
	return Codec{Name: "unknow codec"}
}

func FromName(name string) Codec {
	for _, codec := range allcodec {
		if codec.Name == name {
			return codec
		}
	}
	return Codec{Name: "unknow codec"}
}
