package inventory

import "fmt"

var (
	ErrInsufficientCoins    = fmt.Errorf("insufficient coins")
	ErrInsufficientBonus    = fmt.Errorf("insufficient bonus")
	ErrInvalidDecimalString = fmt.Errorf("invalid decimal string, unable to convert to decimal value")
)

func (ih *InventoryHandler) AddCoins(amount int64) error {
	ih.i.Coins += amount
	return ih.m.Inventories.UpdateCoins(ih.i)
}

// Will error if the total amount of coins is less than the amount to remove
func (ih *InventoryHandler) RemoveCoins(amount int64) error {
	if ih.i.Coins < amount {
		return ErrInsufficientCoins
	}
	ih.i.Coins -= amount
	return ih.m.Inventories.UpdateCoins(ih.i)
}

func (ih *InventoryHandler) SetCoins(amount int64) error {
	ih.i.Coins = amount
	return ih.m.Inventories.UpdateCoins(ih.i)
}

func (ih *InventoryHandler) AddCoinBonus(decStr string) error {
	amount, err := stringDecimalToFloat32(decStr)
	if err != nil {
		return err
	}
	ih.i.CoinBonus += amount
	return ih.m.Inventories.UpdateCoinBonus(ih.i)
}

func (ih *InventoryHandler) RemoveCoinBonus(decStr string) error {
	amount, err := stringDecimalToFloat32(decStr)
	if err != nil {
		return err
	}
	if ih.i.CoinBonus < amount {
		return ErrInsufficientBonus
	}
	ih.i.CoinBonus -= amount
	return ih.m.Inventories.UpdateCoinBonus(ih.i)
}

func (ih *InventoryHandler) SetCoinBonus(decStr string) error {
	amount, err := stringDecimalToFloat32(decStr)
	if err != nil {
		return err
	}
	ih.i.CoinBonus = amount
	return ih.m.Inventories.UpdateCoinBonus(ih.i)
}

// Takes in a string decimal and converts it to a float32 (rounding to 2nd decimal place and carrying the 3rd decimal place)
// eg. 1.234 -> 1.23 and 1.235 -> 1.24
func stringDecimalToFloat32(s string) (float32, error) {
	var f float32
	_, err := fmt.Sscanf(s, "%f", &f)
	if err != nil {
		return 0, ErrInvalidDecimalString
	}

	// WARNING: We boutta do some witchcraft that can and will become a footgun
	f *= 100
	f += 0.5
	f = float32(int(f))
	f /= 100

	return f, nil
}
