package dto

type CreateCharacterRequest struct {
	Name         string `json:"name" validate:"required,min=1,max=60" example:"Furina"`
	Rarity       int    `json:"rarity" validate:"required,oneof=4 5" example:"5"`
	Region       string `json:"region" validate:"required,oneof=Mondstadt Liyue Inazuma Sumeru Fontaine Natlan Snezhnaya Nod-Krai Unknown" example:"Fontaine"`
	Element      string `json:"element" validate:"required,oneof=Anemo Geo Electro Dendro Hydro Pyro Cryo" example:"Hydro"`
	WeaponType   string `json:"weaponType" validate:"required,oneof=Sword Claymore Polearm Bow Catalyst" example:"Sword"`
	NateRelease  string `json:"dateRelease" validate:"required,datetime=2006-01-02" example:"2024-09-17"`
	PatchRelease string `json:"patchRelease" validate:"required,min=1,max=20" example:"3.5"`
	ImagePath    string `json:"imagePath" validate:"required,min=1,max=255" example:"/images/characters/furina.png"`
}

type FetchCharactersResponse struct {
	
}