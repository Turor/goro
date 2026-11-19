package network

import (
	"encoding/binary"
	"fmt"
)

// Char-server PIN packets (Hercules pincode.c / char.c).
const (
	PacketHCPincodeLoginState  uint16 = 0x08b9
	PacketHCPincodeLoginState2 uint16 = 0x0ae9
	PacketHCPincodeMakeState   uint16 = 0x08bb
	PacketHCPincodeEditState   uint16 = 0x08bf
	PacketCZPincodeCheck       uint16 = 0x08b8
	PacketCZPincodeFirst       uint16 = 0x08ba
	PacketCZPincodeChange      uint16 = 0x08be
	PacketCZPincodeWindow      uint16 = 0x08c5

	PincodeLoginOK         uint16 = 0
	PincodeLoginAsk        uint16 = 1
	PincodeLoginNotSet     uint16 = 2
	PincodeLoginExpired    uint16 = 3
	PincodeLoginRestrictPW uint16 = 5
	PincodeLoginWrong      uint16 = 8

	PincodeLoginFlagLocked uint16 = 0
	PincodeLoginFlagWrong  uint16 = 2

	PincodeMakeSuccess    uint16 = 0
	PincodeMakeRestrictPW uint16 = 2
	PincodeMakeFailed     uint16 = 4

	PincodeEditSuccess    uint16 = 0
	PincodeEditFailed     uint16 = 1
	PincodeEditRestrictPW uint16 = 2

	pincodeMultiplier uint32 = 0x3498
	pincodeBaseSeed   uint32 = 0x881234
	pincodeDigits            = 4
)

type PincodeLoginState struct {
	Seed      uint32
	AccountID uint32
	State     uint16
	Flag      uint16
	HasFlag   bool
}

type PincodeSeedState struct {
	State uint16
	Seed  uint32
}

func ParsePincodeLoginState(p Packet) (PincodeLoginState, bool, error) {
	switch p.ID {
	case PacketHCPincodeLoginState:
		if len(p.Data) != 12 {
			return PincodeLoginState{}, false, fmt.Errorf("HC_PINCODE_LOGIN 0x08B9: invalid length %d", len(p.Data))
		}
		return PincodeLoginState{
			Seed:      binary.LittleEndian.Uint32(p.Data[2:6]),
			AccountID: binary.LittleEndian.Uint32(p.Data[6:10]),
			State:     binary.LittleEndian.Uint16(p.Data[10:12]),
		}, true, nil
	case PacketHCPincodeLoginState2:
		if len(p.Data) < 13 {
			return PincodeLoginState{}, false, fmt.Errorf("HC_PINCODE_LOGIN2 0x0AE9: invalid length %d", len(p.Data))
		}
		flag := uint16(p.Data[12])
		if len(p.Data) >= 14 {
			flag = binary.LittleEndian.Uint16(p.Data[12:14])
		}
		return PincodeLoginState{
			Seed:      binary.LittleEndian.Uint32(p.Data[2:6]),
			AccountID: binary.LittleEndian.Uint32(p.Data[6:10]),
			State:     binary.LittleEndian.Uint16(p.Data[10:12]),
			Flag:      flag,
			HasFlag:   true,
		}, true, nil
	default:
		return PincodeLoginState{}, false, nil
	}
}

func ParsePincodeMakeState(p Packet) (PincodeSeedState, bool, error) {
	if p.ID != PacketHCPincodeMakeState {
		return PincodeSeedState{}, false, nil
	}
	if len(p.Data) != 8 {
		return PincodeSeedState{}, false, fmt.Errorf("HC_PINCODE_MAKE 0x08BB: invalid length %d", len(p.Data))
	}
	return PincodeSeedState{
		State: binary.LittleEndian.Uint16(p.Data[2:4]),
		Seed:  binary.LittleEndian.Uint32(p.Data[4:8]),
	}, true, nil
}

func ParsePincodeEditState(p Packet) (PincodeSeedState, bool, error) {
	if p.ID != PacketHCPincodeEditState {
		return PincodeSeedState{}, false, nil
	}
	if len(p.Data) != 8 {
		return PincodeSeedState{}, false, fmt.Errorf("HC_PINCODE_EDIT 0x08BF: invalid length %d", len(p.Data))
	}
	return PincodeSeedState{
		State: binary.LittleEndian.Uint16(p.Data[2:4]),
		Seed:  binary.LittleEndian.Uint32(p.Data[4:8]),
	}, true, nil
}

// EncryptPincode maps plaintext digits through the Gravity keypad shuffle
// so Hercules pincode_decrypt recovers the original PIN.
func EncryptPincode(seed uint32, pin string) string {
	tab := pincodeShuffle(seed)
	keypad := [10]byte{}
	for i := 0; i < 10; i++ {
		keypad[tab[i]] = byte(i)
	}
	out := make([]byte, 0, len(pin))
	for i := 0; i < len(pin); i++ {
		d := pin[i] - '0'
		if d > 9 {
			d = 0
		}
		out = append(out, '0'+keypad[d])
	}
	return string(out)
}

func decryptPincode(seed uint32, pin string) string {
	tab := pincodeShuffle(seed)
	out := make([]byte, 0, len(pin))
	for i := 0; i < len(pin); i++ {
		d := pin[i] - '0'
		if d > 9 {
			d = 0
		}
		out = append(out, '0'+tab[d])
	}
	return string(out)
}

func pincodeShuffle(seed uint32) [10]byte {
	var tab [10]byte
	for i := 0; i < 10; i++ {
		tab[i] = byte(i)
	}
	userSeed := seed
	for i := 1; i < 10; i++ {
		userSeed = pincodeBaseSeed + userSeed*pincodeMultiplier
		pos := int(userSeed % uint32(i+1))
		if i != pos {
			tab[i], tab[pos] = tab[pos], tab[i]
		}
	}
	return tab
}

func padPincode(pin string) [4]byte {
	var out [4]byte
	copy(out[:], []byte("0000"))
	n := len(pin)
	if n > pincodeDigits {
		n = pincodeDigits
	}
	copy(out[:], pin[:n])
	return out
}

func BuildPincodeCheckPacket(accountID uint32, encryptedPIN string) []byte {
	data := make([]byte, 10)
	binary.LittleEndian.PutUint16(data[0:2], PacketCZPincodeCheck)
	binary.LittleEndian.PutUint32(data[2:6], accountID)
	pin := padPincode(encryptedPIN)
	copy(data[6:10], pin[:])
	return data
}

func BuildPincodeFirstPacket(accountID uint32, encryptedPIN string) []byte {
	data := make([]byte, 10)
	binary.LittleEndian.PutUint16(data[0:2], PacketCZPincodeFirst)
	binary.LittleEndian.PutUint32(data[2:6], accountID)
	pin := padPincode(encryptedPIN)
	copy(data[6:10], pin[:])
	return data
}

func BuildPincodeChangePacket(accountID uint32, encryptedOld, encryptedNew string) []byte {
	data := make([]byte, 14)
	binary.LittleEndian.PutUint16(data[0:2], PacketCZPincodeChange)
	binary.LittleEndian.PutUint32(data[2:6], accountID)
	oldPin := padPincode(encryptedOld)
	newPin := padPincode(encryptedNew)
	copy(data[6:10], oldPin[:])
	copy(data[10:14], newPin[:])
	return data
}

func (c *Client) SendPincodeCheck(accountID uint32, seed uint32, pin string) error {
	return c.Send(BuildPincodeCheckPacket(accountID, EncryptPincode(seed, pin)))
}

func (c *Client) SendPincodeFirst(accountID uint32, seed uint32, pin string) error {
	return c.Send(BuildPincodeFirstPacket(accountID, EncryptPincode(seed, pin)))
}

func (c *Client) SendPincodeChange(accountID uint32, seed uint32, oldPIN, newPIN string) error {
	return c.Send(BuildPincodeChangePacket(accountID, EncryptPincode(seed, oldPIN), EncryptPincode(seed, newPIN)))
}
