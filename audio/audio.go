package audio

// typedef unsigned char Uint8;
// void SineWave(void *userdata, Uint8 *stream, int len);
import "C"
import (
	"math"
	"reflect"
	"unsafe"

	"github.com/cockroachdb/errors"
	"github.com/veandco/go-sdl2/sdl"
)

const (
	DefaultFrequency = 16000
	DefaultFormat    = sdl.AUDIO_S16
	DefaultChannels  = 2
	DefaultSamples   = 512

	toneHz = 80
	dPhase = 2 * math.Pi * toneHz / DefaultSamples
)

//export SineWave
func SineWave(userdata unsafe.Pointer, stream *C.Uint8, length C.int) {
	n := int(length) / 2
	hdr := reflect.SliceHeader{Data: uintptr(unsafe.Pointer(stream)), Len: n, Cap: n}
	buf := *(*[]C.ushort)(unsafe.Pointer(&hdr))

	var phase float64
	for i := 0; i < n; i++ {
		phase += dPhase
		sample := C.ushort((math.Sin(phase) + 0.999999) * 32768)
		buf[i] = sample
	}
}

func InitAudioDevice() (sdl.AudioDeviceID, error) {
	// Specify the configuration for our default playback device
	spec := sdl.AudioSpec{
		Freq:     DefaultFrequency,
		Format:   DefaultFormat,
		Channels: DefaultChannels,
		Samples:  DefaultSamples,
		Callback: sdl.AudioCallback(C.SineWave),
	}

	// Open default playback device
	var dev sdl.AudioDeviceID
	dev, err := sdl.OpenAudioDevice("", false, &spec, nil, 0)
	if err != nil {
		return 0, errors.Wrap(err, "error opening audio playback device")
	}

	return dev, nil
}

func New() (*Audio, error) {
	// Specify the configuration for our default playback device
	spec := sdl.AudioSpec{
		Freq:     DefaultFrequency,
		Format:   DefaultFormat,
		Channels: DefaultChannels,
		Samples:  DefaultSamples,
		Callback: sdl.AudioCallback(C.SineWave),
	}

	// Open default playback device
	dev, err := sdl.OpenAudioDevice("", false, &spec, nil, 0)
	if err != nil {
		return nil, errors.Wrap(err, "error opening audio playback device")
	}

	return &Audio{device: dev}, nil
}

type Audio struct {
	device sdl.AudioDeviceID
}

func (a *Audio) Beep(on bool) error {
	sdl.PauseAudioDevice(a.device, !on)
	return nil
}

func (a *Audio) Close() {
	sdl.CloseAudioDevice(a.device)
}
