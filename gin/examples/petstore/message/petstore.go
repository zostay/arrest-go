package message

type Pet struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
	Tag  string `json:"tag"`
}

type Pets []Pet

type Error struct {
	Code    int32  `json:"code"`
	Message string `json:"message"`
}

func ListPets(limit int32) (Pets, error) {
	// This is where you would put your implementation of ListPets
	return nil, nil
}

func CreatePets(pet Pet) error {
	// This is where you would put your implementation of CreatePets
	return nil
}

func ShowByPetID(petID string) (*Pet, error) {
	// This is where you would put your implementation of ShowByPetID
	return nil, nil
}
