package db

import (
	"errors"
	"log"
	"service4/pkg/db"
	"service4/pkg/entity"

	"gorm.io/gorm"
)

var DB *gorm.DB
var err error

func init() {
	DB, err = db.ConnectToDB()
	if err != nil {
		log.Fatal(err)
	}
}

func GetAllOrders(userId, offset, limit int) ([]entity.Order, error) {
	var order []entity.Order
	result := DB.Offset(offset).Limit(limit).Where("user_id=?", userId).Find(&order)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, errors.New("Order not found")
		}
		return nil, errors.New("Order not found")
	}
	return order, nil
}

func Create(userid int) (*entity.Cart, error) {
	cart := &entity.Cart{
		UserId: userid,
	}
	if err := DB.Create(cart).Error; err != nil {
		return nil, err
	}
	return cart, nil

}

func UpdateCart(cart *entity.Cart) error {
	return DB.Where("user_id = ?", cart.UserId).Save(&cart).Error
}

func GetByUserID(userid int) (*entity.Cart, error) {
	var cart entity.Cart
	result := DB.Where("user_id=?", userid).First(&cart)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, errors.New("cart not found")
		}
		return nil, errors.New("cart not found")
	}
	return &cart, nil
}

func GetCartById(userId int) (*entity.Cart, error) {
	var cart entity.Cart
	result := DB.Where("user_id=?", userId).First(&cart)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, result.Error
		}
		return nil, result.Error
	}
	return &cart, nil
}

func CreateCartItem(cartItem *entity.CartItem) error {
	if err := DB.Create(cartItem).Error; err != nil {
		return err
	}
	return nil
}

func UpdateCartItem(cartItem *entity.CartItem) error {
	return DB.Save(cartItem).Error
}

func RemoveCartItem(cartItem *entity.CartItem) error {
	return DB.Where("product_name=?", cartItem.ProductName).Delete(&cartItem).Error
}

func GetByName(productName string, cartId int) (*entity.CartItem, error) {
	var cartItem entity.CartItem
	result := DB.Where("product_name = ? AND cart_id = ?", productName, cartId).First(&cartItem)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, gorm.ErrRecordNotFound
		}
		return nil, result.Error
	}
	return &cartItem, nil
}

func GetAllCartItems(cartId int) ([]entity.CartItem, error) {
	var cartItems []entity.CartItem
	result := DB.Where("cart_id=?", cartId).Find(&cartItems)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, result.Error
		}
		return nil, result.Error
	}
	return cartItems, nil
}

func GetByType(userId int, addressType string) (*entity.Address, error) {
	var address entity.Address
	result := DB.Where(&entity.Address{UserId: userId, Type: addressType}).First(&address)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, errors.New("User address not found - add address")
		}
		return nil, result.Error
	}
	return &address, nil
}

func AddApparelToWishlist(apparel *entity.Wishlist) error {
	if err := DB.Create(apparel).Error; err != nil {
		return err
	}
	return nil
}

func GetApparelFromWishlist(category string, id int, userId int) (bool, error) {
	var apparel entity.Wishlist
	result := DB.Where(&entity.Wishlist{UserId: userId, Category: category, ProductId: id}).First(&apparel)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return false, nil
		}
		return false, errors.New("Error finding apparel")
	}
	return true, nil
}

func GetWishlist(userId int) (*[]entity.Wishlist, error) {
	var wishlist []entity.Wishlist
	result := DB.Where("user_id=?", userId).Find(&wishlist)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, result.Error
		}
		return nil, result.Error
	}
	return &wishlist, nil
}

func RemoveFromWishlist(category string, id, userId int) error {
	product := entity.Wishlist{
		ProductId: id,
		UserId:    userId,
		Category:  category,
	}
	return DB.Where("user_id=?", userId).Delete(&product).Error
}
