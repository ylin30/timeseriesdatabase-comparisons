package common

import (
	"io"
	"strconv"
)

type SerializerOpenTSDBPlain struct {
}

func NewSerializerOpenTSDBPlain() *SerializerOpenTSDBPlain {
	return &SerializerOpenTSDBPlain{}
}

// SerializeOpenTSDBBulk writes Point data to the given writer, conforming to
// the OpenTSDB bulk load protocol (the /api/put endpoint). Note that no line
// has a trailing comma. Downstream programs are responsible for creating
// batches for POSTing using an array.
// We use only millisecond-precision timestamps.
//
// N.B. OpenTSDB only supports millisecond or second resolution timestamps.
// N.B. OpenTSDB millisecond timestamps must be 13 digits long.
// N.B. OpenTSDB only supports floating-point field values.
//
// This function writes JSON lines that looks like:
// { <metric>, <timestamp>, <value>, <tags> }
//
// For example:
// put cpu.usage_user 1451606400000 99.5170917755353770 hostname=host_01 region=ap-southeast-2 datacenter=ap-southeast-2a
// 
// The corresponding Json format is: 
//{ "metric": "cpu.usage_user", "timestamp": 14516064000000, "value": 99.5170917755353770, "tags": { "hostname": "host_01", "region": "ap-southeast-2", "datacenter": "ap-southeast-2a" } }
func (m *SerializerOpenTSDBPlain) SerializePoint(w io.Writer, p *Point) (err error) {
	for i := 0; i < len(p.FieldKeys); i++ {
		var value float64
		switch x := p.FieldValues[i].(type) {
		case int:
			value = float64(x)
		case int64:
			value = float64(x)
		case float32:
			value = float64(x)
		case float64:
			value = x
		default:
			panic("bad numeric value for OpenTSDB serialization")
		}

		buf := scratchBufPool.Get().([]byte)
		buf = append(buf, []byte(`put `)...)
		buf = append(buf, p.MeasurementName...)
		buf = append(buf, '.')
		buf = append(buf, p.FieldKeys[i]...)
		buf = append(buf, ' ')
		buf = strconv.AppendInt(buf, p.Timestamp.UTC().UnixNano()/1e6, 10)
		buf = append(buf, ' ')
		buf = strconv.AppendFloat(buf, value, 'f', 16, 64)
		for i := 0; i < len(p.TagKeys); i++ {
		//for i := 0; i < 2; i++ {
			buf = append(buf, ' ')
			buf = append(buf, p.TagKeys[i]...)
			buf = append(buf, '=')
			buf = append(buf, p.TagValues[i]...)
		}
		buf = append(buf, "\n"...)
		_, err = w.Write(buf)

		buf = buf[:0]
		scratchBufPool.Put(buf)
	}

	return nil
}

func (s *SerializerOpenTSDBPlain) SerializeSize(w io.Writer, points int64, values int64) error {
	//return serializeSizeInText(w, points, values)
	return nil
}
