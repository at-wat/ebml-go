package ebml

// Type represents EBML Element data type
type Type int

// EBML Element data types
const (
	TypeMaster = iota
	TypeInt
	TypeUInt
	TypeDate
	TypeFloat
	TypeBinary
	TypeString
	TypeBlock
)

func (t Type) String() string {
	switch t {
	case TypeMaster:
		return "Master"
	case TypeInt:
		return "Int"
	case TypeUInt:
		return "UInt"
	case TypeDate:
		return "Date"
	case TypeFloat:
		return "Float"
	case TypeBinary:
		return "Binary"
	case TypeString:
		return "String"
	case TypeBlock:
		return "Block"
	default:
		return "Unknown type"
	}
}
