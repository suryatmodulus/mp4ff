package mp4

import (
	"encoding/binary"
	"io"
	"io/ioutil"
)

// StssBox - Sync Sample Box (stss - optional)
//
// Contained in : Sample Table box (stbl)
//
// This lists all sync samples (key frames for video tracks) in the data. If absent, all samples are sync samples.
type StssBox struct {
	Version      byte
	Flags        uint32
	SampleNumber []uint32
	lookUp       map[uint32]bool // Used for optimization
}

// DecodeStss - box-specific decode
func DecodeStss(hdr *boxHeader, startPos uint64, r io.Reader) (Box, error) {
	data, err := ioutil.ReadAll(r)
	if err != nil {
		return nil, err
	}
	versionAndFlags := binary.BigEndian.Uint32(data[0:4])
	b := &StssBox{
		Version:      byte(versionAndFlags >> 24),
		Flags:        versionAndFlags & flagsMask,
		SampleNumber: []uint32{},
	}
	ec := binary.BigEndian.Uint32(data[4:8])
	for i := 0; i < int(ec); i++ {
		sample := binary.BigEndian.Uint32(data[(8 + 4*i):(12 + 4*i)])
		b.SampleNumber = append(b.SampleNumber, sample)
	}
	return b, nil
}

// Type - box-specfic type
func (b *StssBox) Type() string {
	return "stss"
}

// Size - box-specfic size
func (b *StssBox) Size() uint64 {
	return uint64(boxHeaderSize + 8 + len(b.SampleNumber)*4)
}

// IsSyncSample - check if a sample is a sync sample
func (b *StssBox) IsSyncSample(sampleNr uint32) (isSync bool) {
	if b.lookUp == nil {
		b.lookUp = make(map[uint32]bool)
		for _, i := range b.SampleNumber {
			b.lookUp[i] = true
		}
	}
	_, isSync = b.lookUp[sampleNr]
	return
}

// Encode - box-specific encode
func (b *StssBox) Encode(w io.Writer) error {
	err := EncodeHeader(b, w)
	if err != nil {
		return err
	}
	buf := makebuf(b)
	sw := NewSliceWriter(buf)
	versionAndFlags := (uint32(b.Version) << 24) + b.Flags
	sw.WriteUint32(versionAndFlags)
	sw.WriteUint32(uint32(len(b.SampleNumber)))
	for i := range b.SampleNumber {
		sw.WriteUint32(b.SampleNumber[i])
	}
	_, err = w.Write(buf)
	return err
}

func (s *StssBox) Dump(w io.Writer, specificBoxLevels, indent, indentStep string) error {
	bd := newBoxDumper(w, indent, s, int(s.Version))
	// TODO. Add more details to stss dump
	return bd.err
}
