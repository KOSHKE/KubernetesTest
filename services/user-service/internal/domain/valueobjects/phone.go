package valueobjects

type Phone struct {
	value string
}

func NewPhone(phone string) Phone {
	return Phone{value: phone}
}

func (p Phone) Value() string {
	return p.value
}

func (p Phone) String() string {
	return p.value
}

func (p Phone) Equals(other Phone) bool {
	return p.value == other.value
}
