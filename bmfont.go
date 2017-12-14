package bmfont

import (
	"encoding/binary"
	"fmt"
	"strings"

	"golang.org/x/text/encoding"
	"golang.org/x/text/encoding/charmap"
)

var Encoding encoding.Encoding = charmap.Windows1252

const (
	INFO_BITFIELD_SMOOTH  = 1 << (7 - iota) // Set if smoothing was turned on
	INFO_BITFIELD_UNICODE                   // Set if it is the unicode charset
	INFO_BITFIELD_ITALIC                    // The font is italic
	INFO_BITFIELD_BOLD                      // The font is bold
	INFO_BITFIELD_FIXED_HEIGHT
)

const (
	COMMON_BITFIELD_PACKED = 1
)

const (
	BLOCK_TYPE_INFO          = 1
	BLOCK_TYPE_COMMON        = 2
	BLOCK_TYPE_PAGES         = 3
	BLOCK_TYPE_CHARS         = 4
	BLOCK_TYPE_KERNING_PAIRS = 5
)

type Info struct {
	FontSize     int16  // The size of the true type font
	BitField     uint8  // Use INFO_BITFIELD_ consts
	CharSet      uint8  // The name of the OEM charset used (when not unicode)
	StretchH     uint16 // The font height stretch in percentage. 100% means no stretch
	Aa           uint8  // The supersampling level used. 1 means no supersampling was used
	PaddingUp    uint8  // The padding for each character (up, right, down, left)
	PaddingRight uint8
	PaddingDown  uint8
	PaddingLeft  uint8
	SpacingHoriz uint8 // The spacing for each character (horizontal, vertical)
	SpacingVert  uint8
	Outline      uint8  // The outline thickness for the characters
	FontName     string // This is the name of the true type font
}

func (i *Info) fromBinary(b []byte) error {
	i.FontSize = int16(binary.LittleEndian.Uint16(b[0:2]))
	i.BitField = b[2]
	i.CharSet = b[3]
	i.StretchH = binary.LittleEndian.Uint16(b[4:6])
	i.Aa = b[6]
	i.PaddingUp = b[7]
	i.PaddingRight = b[8]
	i.PaddingDown = b[9]
	i.PaddingLeft = b[10]
	i.SpacingHoriz = b[11]
	i.SpacingVert = b[12]
	i.Outline = b[13]

	fontBuf := make([]byte, ((len(b)-14)*5)/2)
	if nDst, _, err := Encoding.NewDecoder().Transform(fontBuf, b[14:len(b)], true); err != nil {
		return fmt.Errorf("Error parsing info section font name: %v", err)
	} else {
		i.FontName = string(fontBuf[:nDst])
	}
	return nil
}

type Common struct {
	LineHeight uint16
	Base       uint16
	ScaleW     uint16
	ScaleH     uint16
	Pages      uint16
	BitField   byte
	AlphaChnl  byte
	RedChnl    byte
	GreenChnl  byte
	BlueChnl   byte
}

func (c *Common) fromBinary(b []byte) error {
	c.LineHeight = binary.LittleEndian.Uint16(b[0:2])
	c.Base = binary.LittleEndian.Uint16(b[2:4])
	c.ScaleW = binary.LittleEndian.Uint16(b[4:6])
	c.ScaleH = binary.LittleEndian.Uint16(b[6:8])
	c.Pages = binary.LittleEndian.Uint16(b[8:10])
	c.BitField = b[10]
	c.AlphaChnl = b[11]
	c.RedChnl = b[12]
	c.GreenChnl = b[13]
	c.BlueChnl = b[14]
	return nil
}

type Char struct {
	Id       uint32
	X        uint16
	Y        uint16
	Width    uint16
	Height   uint16
	Xoffset  int16
	Yoffset  int16
	Xadvance int16
	Page     uint8
	Chnl     uint8
}

func (c *Char) fromBinary(b []byte) error {
	c.Id = binary.LittleEndian.Uint32(b[0:4])
	c.X = binary.LittleEndian.Uint16(b[4:6])
	c.Y = binary.LittleEndian.Uint16(b[6:8])
	c.Width = binary.LittleEndian.Uint16(b[8:10])
	c.Height = binary.LittleEndian.Uint16(b[10:12])
	c.Xoffset = int16(binary.LittleEndian.Uint16(b[12:14]))
	c.Yoffset = int16(binary.LittleEndian.Uint16(b[14:16]))
	c.Xadvance = int16(binary.LittleEndian.Uint16(b[16:18]))
	c.Page = b[18]
	c.Chnl = b[19]
	return nil
}

type KerningPair struct {
	First  uint32
	Second uint32
	Amount uint16
}

func (kp *KerningPair) fromBinary(b []byte) error {
	kp.First = binary.LittleEndian.Uint32(b[0:4])
	kp.Second = binary.LittleEndian.Uint32(b[4:8])
	kp.Amount = binary.LittleEndian.Uint16(b[8:10])
	return nil
}

type Font struct {
	Info         *Info
	Common       *Common
	Pages        []string
	Chars        []Char
	KerningPairs []KerningPair
}

func NewFont() *Font {
	return &Font{}
}

func (f *Font) FromBuffer(b []byte) error {
	if b[0] != 'B' || b[1] != 'M' || b[2] != 'F' {
		return fmt.Errorf("Invalid identifier %v", b[:3])
	}

	if b[3] != 3 {
		return fmt.Errorf("Unsupported version %v", b[4])
	}

	floatBuffer := b[4:]
	for len(floatBuffer) > 4 {
		blockId := floatBuffer[0]
		blockLenght := binary.LittleEndian.Uint32(floatBuffer[1:5])
		blockData := floatBuffer[5 : 5+blockLenght]

		switch blockId {
		case BLOCK_TYPE_INFO:
			f.Info = &Info{}
			if err := f.Info.fromBinary(blockData); err != nil {
				return fmt.Errorf("Error parsing info block: %v", err)
			}
		case BLOCK_TYPE_COMMON:
			f.Common = &Common{}
			if err := f.Common.fromBinary(blockData); err != nil {
				return fmt.Errorf("Error parsing common block: %v", err)
			}
		case BLOCK_TYPE_PAGES:
			fontBuf := make([]byte, (blockLenght*5)/2)
			if nDst, _, err := Encoding.NewDecoder().Transform(fontBuf, blockData, false); err != nil {
				return fmt.Errorf("Error parsing pages text: %v")
			} else {
				f.Pages = strings.Split(string(fontBuf[:nDst]), "\x00")
				f.Pages = f.Pages[:len(f.Pages)-1]
			}
		case BLOCK_TYPE_CHARS:
			charsCnt := blockLenght / 20
			f.Chars = make([]Char, charsCnt)
			for i := range f.Chars {
				if err := f.Chars[i].fromBinary(blockData[i*20 : i*20+20]); err != nil {
					return fmt.Errorf("Error parsing char %v: %v", i, err)
				}
			}
		case BLOCK_TYPE_KERNING_PAIRS:
			kerningPairsCnt := blockLenght / 10
			f.KerningPairs = make([]KerningPair, kerningPairsCnt)
			for i := range f.KerningPairs {
				if err := f.KerningPairs[i].fromBinary(blockData[i*10 : i*10+10]); err != nil {
					return fmt.Errorf("Error parsing kerning pair %v: %v", i, err)
				}
			}
		}

		floatBuffer = floatBuffer[5+blockLenght:]
	}
	return nil
}

func NewFontFromBuf(b []byte) (*Font, error) {
	f := NewFont()
	return f, f.FromBuffer(b)
}
