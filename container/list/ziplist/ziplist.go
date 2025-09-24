package ziplist

import (
	"encoding/binary"
)

type zlentryFlag uint32

const (
	zleMaxByte    = 1<<6 - 1
	zleMaxInt16   = 1<<14 - 1
	zleMaxInt32   = 1<<30 - 1
	zleRaw        = 0xC0
	zleValue      = 0x3F
	zleInt        = 0x00
	zleInt16      = 0x01
	zleInt32      = 0x02
	zleInt64      = 0x03
	zleInt16Size  = 2
	zleInt32Size  = 4
	zleInt64Size  = 8
	zleRawSizeMax = 1<<6 - 1
	zleRawSixBit  = 0x3F
	zleEnd        = 0xFF
)

const (
	ZIP_LIST_END = 255

	ZIP_INT_16B = 0xc1
	ZIP_INT_32B = 0xc2
	ZIP_INT_64B = 0xc3

	ZIP_STR_06B = 0x00
	ZIP_STR_14B = 0x40
	ZIP_STR_32B = 0x80
)

const (
	ZIPLIST_ENCODING_RAW = 0x00
	ZIPLIST_ENCODING_INT = 0x01
)

type zlentry struct {
	prelen uint32
	// 编码方式，可以是ZIPLIST_ENCODING_RAW或ZIPLIST_ENCODING_INT
	encoding byte
	// 数据长度，如果是int类型则为8，否则为实际数据长度
	length uint32
}

// 底层是一个环形数组，
type Ziplist struct {
	// 内存池，用于分配和释放内存
	bytes []byte
	// 元素数量
	length uint32
	// 尾节点偏移量
	tail uint32
	// 尾节点偏移量
	head uint32
}

// 新建一个ziplist
func New() *Ziplist {
	z := &Ziplist{}
	z.bytes = make([]byte, 0, 1024)
	return z
}

// 在ziplist中添加一个元素
func (z *Ziplist) push(value []byte) error {
	var prelen uint32
	if z.tail != 0 && z.length > 0 {
		if z.bytes[z.tail+4] == ZIPLIST_ENCODING_RAW {
			prelen = binary.LittleEndian.Uint32(z.bytes[z.tail+5:]) + 9
		} else {
			prelen = 13
		}
	}

	// 写入数据
	binary.LittleEndian.AppendUint32(z.bytes, prelen)
	z.bytes = append(z.bytes, ZIPLIST_ENCODING_RAW)
	binary.LittleEndian.AppendUint32(z.bytes, uint32(len(value)))
	copy(z.bytes[5:], value)

	// 更新tail指针
	if z.tail != 0 && z.length > 0 {
		z.tail += prelen
	}

	// 更新head指针
	if z.length == 0 && z.head == 0 {
		z.head = 0
	}

	// 更新元素数量
	z.length++

	return nil
}

// 在ziplist中添加一个int类型的元素
func (z *Ziplist) pushInt(value int64) error {

	var prelen uint32
	if z.tail != 0 && z.length > 0 {
		if z.bytes[z.tail+4] == ZIPLIST_ENCODING_RAW {
			prelen = binary.LittleEndian.Uint32(z.bytes[z.tail+5:]) + 9
		} else {
			prelen = 13
		}
	}

	// 写入数据
	binary.LittleEndian.AppendUint32(z.bytes, prelen)
	z.bytes = append(z.bytes, ZIPLIST_ENCODING_INT)
	binary.LittleEndian.AppendUint64(z.bytes, uint64(value))

	// 更新tail指针
	if z.tail != 0 && z.length > 0 {
		z.tail += prelen
	}

	// 更新head指针
	if z.length == 0 && z.head == 0 {
		z.head = 0
	}

	// 更新元素数量
	z.length++

	return nil
}
