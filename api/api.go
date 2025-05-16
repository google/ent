package api

const GET_ENTRY_METHOD_ID = "z07773083324874402207"

type GetEntryResponse struct {
	Metadata ObjectMetadata `json:"z11701617426848460867"`
	Mirrors  []Mirror       `json:"z02797662040442636406"`
}

type Mirror struct {
	URL  string `json:"z15006311556098510585"`
	CORS string `json:"z03067985653251929561"`
}

type ObjectMetadata struct {
	Digests     Digests `json:"z00760714168124038847"`
	LengthBytes int64   `json:"z05966774115567221820"`
	ContentType string  `json:"z12467592263966562957"`
}

// Hex encoded strings.
type Digests struct {
	Sha2_256 string `json:"sha2-256"`
	Sha2_512 string `json:"sha2-512"`

	Sha3_256 string `json:"sha3-256"`
	Sha3_384 string `json:"sha3-384"`
	Sha3_512 string `json:"sha3-512"`
}
