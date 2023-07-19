package db

import (
	"errors"
	"log"
	"service4/pkg/db"
	"service4/pkg/entity"
	"time"

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

func GetAddressById(addressId int) (*entity.Address, error) {
	var address entity.Address
	result := DB.Where(&entity.Address{ID: addressId}).First(&address)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, result.Error
	}
	return &address, nil
}

func CreateOrder(order *entity.Order) (int, error) {
	if err := DB.Create(order).Error; err != nil {
		return 0, err
	}
	return int(order.ID), nil
}

func CreateOrderItems(orderItem []entity.OrderItem) error {
	if err := DB.Create(orderItem).Error; err != nil {
		return err
	}
	return nil
}

func CreateInvoice(invoice *entity.Invoice) (*entity.Invoice, error) {
	if err := DB.Create(invoice).Error; err != nil {
		return nil, err
	}
	return invoice, nil
}

func CreateReturn(returnData *entity.Return) error {
	if err := DB.Create(returnData).Error; err != nil {
		return err
	} else {
		return nil
	}
}
func GetReturnByID(returnId int) (*entity.Return, error) {
	var returnData entity.Return
	result := DB.Where("id=?", returnId).First(&returnData)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, errors.New("Return Data not found")
		}
		return nil, errors.New("Return Data not found")
	}
	return &returnData, nil
}
func GetReturnByOrderID(orderId int) (*entity.Return, error) {
	var returnData entity.Return
	result := DB.Where("order_id=?", orderId).First(&returnData)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, errors.New("Return Data not found")
		}
		return nil, errors.New("Return Data not found")
	}
	return &returnData, nil
}
func UpdateReturn(returnData *entity.Return) error {
	return DB.Save(&returnData).Error
}

func UpdateUserWallet(user *entity.User) error {
	return DB.Save(&user).Error
}

func DecreaseProductQuantity(product *entity.Inventory) error {
	existingProduct := &entity.Inventory{}
	err := DB.Where("product_category = ? AND product_id =?", product.ProductCategory, product.ProductId).First(existingProduct).Error
	if err != nil {
		return err
	}
	newQuantity := existingProduct.Quantity - product.Quantity
	err = DB.Model(existingProduct).Update("Quantity", newQuantity).Error
	if err != nil {
		return err
	}
	return nil
}

func RemoveCartItems(cartId int) error {
	var cartItems entity.CartItem
	result := DB.Where("cart_id=?", cartId).Delete(&cartItems)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return result.Error
		}
		return result.Error
	}
	return nil
}

func GetByRazorId(razorId string) (*entity.Order, error) {
	var order entity.Order
	result := DB.Where("payment_id=?", razorId).First(&order)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, errors.New("Order not found")
		}
		return nil, errors.New("Order not found")
	}
	return &order, nil
}

func Update(order *entity.Order) error {
	return DB.Save(&order).Error
}

func GetByID(orderId int) (*entity.Order, error) {
	var order entity.Order
	result := DB.Where("id=?", orderId).First(&order)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, errors.New("Order not found")
		}
		return nil, errors.New("Order not found")
	}
	return &order, nil
}

func GetUserByID(id int) (*entity.User, error) {
	var user entity.User
	result := DB.First(&user, id)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, result.Error
	}
	return &user, nil
}

func GetByDate(startDate, endDate time.Time) (*entity.SalesReport, error) {
	var Order []entity.Order
	var report entity.SalesReport

	if err := DB.Model(&Order).Where("created_at BETWEEN ? AND ?", startDate, endDate).Select("SUM(total) as total_sales").Scan(&report).Error; err != nil {
		return nil, err
	}

	if err := DB.Model(&Order).Where("created_at BETWEEN ? AND ?", startDate, endDate).Count(&report.TotalOrders).Error; err != nil {
		return nil, err
	}

	if err := DB.Model(&Order).Where("created_at BETWEEN ? AND ?", startDate, endDate).Select("AVG(total) as average_order").Scan(&report).Error; err != nil {
		return nil, err
	}

	if err := DB.Model(&Order).Where("created_at BETWEEN ? AND ?", startDate, endDate).Select("payment, COUNT(payment) as payment_method_cnt").Group("payment").Scan(&report).Error; err != nil {
		return nil, err
	}
	return &report, nil
}

func GetByCategory(category string, startDate, endDate time.Time) (*entity.SalesReport, error) {
	report := &entity.SalesReport{}

	var orderItems []entity.OrderItem
	if err := DB.Where("category = ? AND created_at BETWEEN ? AND ?", category, startDate, endDate).Find(&orderItems).Error; err != nil {
		return nil, err
	}

	totalSales := 0.0
	totalOrders := int64(len(orderItems))

	for _, item := range orderItems {
		totalSales += item.Price * float64(item.Quantity)

	}

	report.TotalSales = totalSales
	report.TotalOrders = totalOrders
	report.AverageOrder = totalSales / float64(totalOrders)

	return report, nil
}

func GetByStatus(offset, limit int, status string) ([]entity.Order, error) {
	var order []entity.Order
	result := DB.Offset(offset).Limit(limit).Where("status=?", status).Find(&order)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, errors.New("Order not found")
		}
		return nil, errors.New("Order not found")
	}
	return order, nil
}
