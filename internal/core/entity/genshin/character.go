package genshin

import (
	"be-ayaka/internal/core/entity"
	"time"
)

type Region string
type Element string
type WeaponType string

const (
	RegionMondstadt Region = "Mondstadt"
	RegionLiyue     Region = "Liyue"
	RegionInazuma   Region = "Inazuma"
	RegionSumeru    Region = "Sumeru"
	RegionFontaine  Region = "Fontaine"
	RegionNatlan    Region = "Natlan"
	RegionSnezhnaya Region = "Snezhnaya"
	RegionNodKrai   Region = "Nod-Krai"
	RegionUnknown   Region = "Unknown"

	ElementAnemo   Element = "Anemo"
	ElementGeo     Element = "Geo"
	ElementElectro Element = "Electro"
	ElementDendro  Element = "Dendro"
	ElementHydro   Element = "Hydro"
	ElementPyro    Element = "Pyro"
	ElementCryo    Element = "Cryo"

	WeaponTypeSword    WeaponType = "Sword"
	WeaponTypeClaymore WeaponType = "Claymore"
	WeaponTypePolearm  WeaponType = "Polearm"
	WeaponTypeBow      WeaponType = "Bow"
	WeaponTypeCatalyst WeaponType = "Catalyst"
)

type GiCharacter struct {
	entity.BaseEntity
	Name         string     `gorm:"unique;not null;type:varchar(60)"`
	Rarity       int        `gorm:"not null;type:int;check:rarity IN (4, 5)"`
	Region       Region     `gorm:"not null;type:varchar(60);check:region IN ('Mondstadt', 'Liyue', 'Inazuma', 'Sumeru', 'Fontaine', 'Natlan', 'Snezhnaya', 'Nod-Krai', 'Unknown')"`
	Element      Element    `gorm:"not null;type:varchar(60);check:element IN ('Anemo', 'Geo', 'Electro', 'Dendro', 'Hydro', 'Pyro', 'Cryo')"`
	WeaponType   WeaponType `gorm:"not null;type:varchar(60);check:weapon_type IN ('Sword', 'Claymore', 'Polearm', 'Bow', 'Catalyst')"`
	DateRelease  time.Time  `gorm:"not null;type:date"`
	PatchRelease string     `gorm:"not null;type:varchar(20)"`
	ImagePath    string     `gorm:"not null;type:varchar(255)"`
}

type GiCharacterMaterial struct{}

type GiCharacterAscensionCost struct{}

type GiCharacterTalentCost struct{}
