package fpk

import "encoding/json"

type Reference struct {
	FilePath String
}

const ReferenceSize = 4 * 4 // FilePath size

type rjs struct {
	FilePath string `json:"filePath"`
}

func (r *Reference) MarshalJSON() ([]byte, error) {
	rj := rjs{
		FilePath: r.FilePath.Data,
	}

	return json.Marshal(rj)
}

func (r *Reference) UnmarshalJSON(bytes []byte) error {
	rj := rjs{}
	if err := json.Unmarshal(bytes, &rj); err != nil {
		return err
	}

	r.FilePath.Data = rj.FilePath
	return nil
}
