package provider

import (
	"embed"
	"fmt"
	"math"
	"strconv"

	"go.uber.org/atomic"

	"github.com/livekit/protocol/livekit"
)

const (
	h264Codec = "h264"
	vp8Codec  = "vp8"
)

type VideoSpec struct {
	Codec  string
	Prefix string
	Height int
	Width  int
	Kbps   int
	Fps    int
}

func (v *VideoSpec) Name() string {
	ext := "h264"
	if v.Codec == vp8Codec {
		ext = "ivf"
	}
	size := strconv.Itoa(v.Height)
	if v.Height > v.Width {
		size = fmt.Sprintf("p%d", v.Width)
	}
	return fmt.Sprintf("resources/%s_%s_%d.%s", v.Prefix, size, v.Kbps, ext)
}

func (v *VideoSpec) ToVideoLayer(quality livekit.VideoQuality) *livekit.VideoLayer {
	return &livekit.VideoLayer{
		Quality: quality,
		Height:  uint32(v.Height),
		Width:   uint32(v.Width),
		Bitrate: v.bitrate(),
	}
}

func (v *VideoSpec) bitrate() uint32 {
	return uint32(v.Kbps * 1000)
}

func circlesSpec(width, kbps, fps int) *VideoSpec {
	return &VideoSpec{
		Codec:  h264Codec,
		Prefix: "circles",
		Height: width * 4 / 3,
		Width:  width,
		Kbps:   kbps,
		Fps:    fps,
	}
}

func createSpecs(prefix string, codec string, bitrates ...int) []*VideoSpec {
	var specs []*VideoSpec
	videoFps := []int{
		15, 20, 30,
	}
	for i, b := range bitrates {
		dimMultiple := int(math.Pow(2, float64(i)))
		specs = append(specs, &VideoSpec{
			Prefix: prefix,
			Codec:  codec,
			Kbps:   b,
			Fps:    videoFps[i],
			Height: 180 * dimMultiple,
			Width:  180 * dimMultiple * 16 / 9,
		})
	}
	return specs
}

var (
	//go:embed resources
	res embed.FS

	videoSpecs [][]*VideoSpec
	videoIndex atomic.Int64
	audioNames []string
	audioIndex atomic.Int64
)

func init() {
	videoSpecs = [][]*VideoSpec{
		createSpecs("butterfly", h264Codec, 150, 400, 2000),
		createSpecs("cartoon", h264Codec, 120, 400, 1500),
		createSpecs("crescent", vp8Codec, 150, 600, 2000),
		createSpecs("neon", vp8Codec, 150, 600, 2000),
		createSpecs("tunnel", vp8Codec, 150, 600, 2000),
		{
			circlesSpec(180, 200, 15),
			circlesSpec(360, 700, 20),
			circlesSpec(540, 2000, 30),
		},
	}
	audioNames = []string{
		"change-amelia",
		"change-benjamin",
		"change-elena",
		"change-clint",
		"change-emma",
		"change-ken",
		"change-sophie",
	}
}

func randomVideoSpecsForCodec(videoCodec string) []*VideoSpec {
	filtered := make([][]*VideoSpec, 0)
	for _, specs := range videoSpecs {
		if videoCodec == "" || specs[0].Codec == videoCodec {
			filtered = append(filtered, specs)
		}
	}
	chosen := int(videoIndex.Inc()) % len(filtered)
	return filtered[chosen]
}

func CreateVideoLoopers(resolution string, codecFilter string, simulcast bool) ([]VideoLooper, error) {
	specs := randomVideoSpecsForCodec(codecFilter)
	numToKeep := 0
	switch resolution {
	case "medium":
		numToKeep = 2
	case "low":
		numToKeep = 1
	default:
		numToKeep = 3
	}
	specs = specs[:numToKeep]
	if !simulcast {
		specs = specs[numToKeep-1:]
	}
	loopers := make([]VideoLooper, 0)
	for _, spec := range specs {
		f, err := res.Open(spec.Name())
		if err != nil {
			return nil, err
		}
		defer f.Close()
		if spec.Codec == h264Codec {
			looper, err := NewH264VideoLooper(f, spec)
			if err != nil {
				return nil, err
			}
			loopers = append(loopers, looper)
		} else if spec.Codec == vp8Codec {
			looper, err := NewVP8VideoLooper(f, spec)
			if err != nil {
				return nil, err
			}
			loopers = append(loopers, looper)
		}
	}
	return loopers, nil
}

func CreateAudioLooper() (*OpusAudioLooper, error) {
	chosenName := audioNames[int(audioIndex.Load())%len(audioNames)]
	audioIndex.Inc()
	f, err := res.Open(fmt.Sprintf("resources/%s.ogg", chosenName))
	if err != nil {
		return nil, err
	}
	defer f.Close()
	return NewOpusAudioLooper(f)
}
