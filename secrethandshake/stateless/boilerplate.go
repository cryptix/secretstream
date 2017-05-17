package stateless

import (
	"encoding/hex"
	"fmt"
	"reflect"
	"strings"
)

func stripIfZero(s string) string {
	if s == "0000000000000000000000000000000000000000000000000000000000000000" {
		s = ""
	}
	return s
}

// TODO: only expose in tests?
func (s *State) ToJsonState() *JsonState {

	rpubStr := hex.EncodeToString(s.remotePublic[:])
	rephPubStr := hex.EncodeToString(s.ephKeyRemotePub[:])
	secStr := hex.EncodeToString(s.secret[:])
	shStr := hex.EncodeToString(s.secHash[:])
	sec2Str := hex.EncodeToString(s.secret2[:])
	abobStr := hex.EncodeToString(s.aBob[:])

	// zero value means long sequence of "0000..."
	for _, s := range []*string{
		&rpubStr,
		&rephPubStr,
		&secStr,
		&shStr,
		&sec2Str,
		&abobStr,
	} {
		*s = stripIfZero(*s)
	}

	return &JsonState{
		AppKey: hex.EncodeToString(s.appKey),
		Local: localKey{
			KxPK:      hex.EncodeToString(s.ephKeyPair.Public[:]),
			KxSK:      hex.EncodeToString(s.ephKeyPair.Secret[:]),
			PublicKey: hex.EncodeToString(s.local.Public[:]),
			SecretKey: hex.EncodeToString(s.local.Secret[:]),
			AppMac:    hex.EncodeToString(s.localAppMac),
			Hello:     hex.EncodeToString(s.localHello),
		},
		Remote: remoteKey{
			PublicKey: rpubStr,
			EphPubKey: rephPubStr,
			AppMac:    hex.EncodeToString(s.remoteAppMac),
			Hello:     hex.EncodeToString(s.remoteHello),
		},
		Random:  hex.EncodeToString(s.ephRandBuf.Bytes()),
		Secret:  secStr,
		SecHash: shStr,
		Secret2: sec2Str,
		ABob:    abobStr,
	}
}

// json test vectors > go conversion boilerplate
type localKey struct {
	KxPK      string `mapstructure:"kx_pk"`
	KxSK      string `mapstructure:"kx_sk"`
	PublicKey string `mapstructure:"publicKey"`
	SecretKey string `mapstructure:"secretKey"`
	AppMac    string `mapstructure:"app_mac"`
	Hello     string `mapstructure:"hello"`
}

type remoteKey struct {
	PublicKey string `mapstructure:"publicKey"`
	EphPubKey string `mapstructure:"kx_pk"`
	AppMac    string `mapstructure:"app_mac"`
	Hello     string `mapstructure:"hello"`
}

type JsonState struct {
	AppKey  string    `mapstructure:"app_key"`
	Local   localKey  `mapstructure:"local"`
	Remote  remoteKey `mapstructure:"remote"`
	Seed    string    `mapstructure:"seed"`
	Random  string    `mapstructure:"random"`
	Secret  string    `mapstructure:"secret"`
	SecHash string    `mapstructure:"shash"`
	Secret2 string    `mapstructure:"secret2"`
	ABob    string    `mapstructure:"a_bob"`
}

func InitializeFromJSONState(s JsonState) (*State, error) {
	var localKeyPair Option
	if s.Seed != "" {
		localKeyPair = LocalKeyFromSeed(strings.NewReader(s.Seed))
	} else {
		localKeyPair = LocalKeyFromHex(s.Local.PublicKey, s.Local.SecretKey)
	}
	return Initialize(
		SetAppKey(s.AppKey),
		localKeyPair,
		EphemeralRandFromHex(s.Random),
		RemotePubFromHex(s.Remote.PublicKey),
		func(state *State) error {
			if s.Local.Hello != "" {
				var err error
				state.localHello, err = hex.DecodeString(s.Local.Hello)
				if err != nil {
					return err
				}
			}
			return nil
		},
		// func(state *State) error {
		// 	if s.Remote.Hello != "" {
		// 		d, err := hex.DecodeString(s.Remote.Hello)
		// 		if err != nil {
		// 			return err
		// 		}
		// 		copy(state.remoteHello[:], d)
		// 	}
		// 	return nil
		// },
		func(state *State) error {
			if s.Secret != "" {
				s, err := hex.DecodeString(s.Secret)
				if err != nil {
					return err
				}
				copy(state.secret[:], s)
			}
			return nil
		},
		func(state *State) error {
			if s.Secret2 != "" {
				s2, err := hex.DecodeString(s.Secret2)
				if err != nil {
					return err
				}
				copy(state.secret2[:], s2)
			}
			return nil
		},
		func(state *State) error {
			if s.Remote.EphPubKey != "" {
				r, err := hex.DecodeString(s.Remote.EphPubKey)
				if err != nil {
					return err
				}
				copy(state.ephKeyRemotePub[:], r)
			}
			return nil
		},
		func(state *State) error {
			if s.SecHash != "" {
				var err error
				state.secHash, err = hex.DecodeString(s.SecHash)
				if err != nil {
					return err
				}
			}
			return nil
		},
	)
}

// WIP: DRY for the above
func fill(field, value string) Option {
	return func(s *State) error {
		if value != "" {
			b, err := hex.DecodeString(value)
			if err != nil {
				return err
			}
			t, ok := reflect.TypeOf(*s).FieldByName(field)
			if !ok {
				return fmt.Errorf("field not found")
			}

			fmt.Println("Len:", t.Type.Len())
			const l = 32 // t.Type.Len()

			v := reflect.ValueOf(*s).FieldByName(field).Interface().([l]uint8)
			copy(v[:], b)
		}
		return nil
	}
}
