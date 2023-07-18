package service

import (
	"context"
	"errors"
	"log"
	pb "service4/pb"
	"service4/pkg/entity"
	repo "service4/pkg/repository"

	"github.com/gin-gonic/gin"
	"github.com/razorpay/razorpay-go"
)

type Order struct {
	pb.UnimplementedOrderServer
}

func (s *Order) MyMethod(ctx context.Context, req *pb.Request) (*pb.Response, error) {
	log.Println("Microservice1: MyMethod called")

	result := "Hello, " + req.Data
	return &pb.Response{Result: result}, nil
}

func (s *Order) PlaceOrder(ctx context.Context, req *pb.PlaceOrderRequest) (*pb.PlaceOrderResponse, error) {
	resp := &pb.PlaceOrderResponse{}
	if req.Paymentmethod == "cod" {
		result, err := ExecutePurchaseCod(int(req.Userid), int(req.Addressid))
		if err != nil {
			return nil, err
		} else {
			resp.Invoice.Address = result.AddressType
			resp.Invoice.Orderid = int32(result.OrderId)
			resp.Invoice.Payment = result.Payment
			resp.Invoice.Price = int32(result.Price)
			resp.Invoice.Status = result.Status
			resp.Invoice.Userid = int32(result.UserId)
		}
	} else if req.Paymentmethod == "razorpay" {
		sign, _, err := ExecutePurchaseRazorPay(int(req.Userid), int(req.Addressid))
		if err != nil {
			return nil, err
		} else {
			resp.Razorid = sign
			resp.Result = "Payment Initiated"
		}
	}
	return resp, nil
}
func ExecutePurchaseRazorPay(userId int, address int, c *gin.Context) (string, int, error) {
	var orderItems []entity.OrderItem
	cart, err := repo.GetCartById(userId)
	if err != nil {
		return "", 0, errors.New("Cart  not found")
	}
	cartItems, err1 := repo.GetAllCartItems(int(cart.ID))
	if err1 != nil {
		return "", 0, errors.New("Cart Items  not found")
	}
	userAddress, err := repo.GetAddressById(address)
	if err != nil {
		return "", 0, errors.New("User address  not found")
	}
	client := razorpay.NewClient("rzp_test_O6q2DXJHecJBHI", "MU9PWzkhTBSCkPnxEUOAZdYW")

	data := map[string]interface{}{
		"amount":   int(cart.TotalPrice) * 100,
		"currency": "INR",
		"receipt":  "101",
	}
	body, err := client.Order.Create(data, nil)
	if err != nil {
		return "", 0, errors.New("Payment not initiated")
	}
	razorId, _ := body["id"].(string)
	Total := cart.TotalPrice - float64(cart.OfferPrice)
	order := &entity.Order{
		UserID:        cart.UserId,
		AddressId:     userAddress.ID,
		Total:         Total,
		Status:        "pending",
		PaymentMethod: "razorpay",
		PaymentStatus: "pending",
		PaymentId:     razorId,
	}
	OrderId, err2 := repo.Create(order)
	if err2 != nil {
		return "", 0, errors.New("Order placing failed")
	}
	for _, cartItem := range cartItems {
		orderItem := entity.OrderItem{
			OrderID:   OrderId,
			ProductID: cartItem.ProductId,
			Category:  cartItem.Category,
			Quantity:  cartItem.Quantity,
			Price:     cartItem.Price,
		}
		orderItems = append(orderItems, orderItem)
	}

	err3 := repo.CreateOrderItems(orderItems)
	if err3 != nil {
		return "", 0, errors.New("User cart is empty")
	}
	return razorId, OrderId, nil
}

func ExecutePurchaseCod(userId int, address int) (*entity.Invoice, error) {
	var orderItems []entity.OrderItem
	cart, err := repo.GetCartById(userId)
	if err != nil {
		return nil, errors.New("Cart  not found")
	}
	cartItems, err1 := repo.GetAllCartItems(int(cart.ID))
	if err1 != nil {
		return nil, errors.New("Cart Items  not found")
	}
	userAddress, err := repo.GetAddressById(address)
	if err != nil {
		return nil, errors.New("User address  not found")
	}
	Total := cart.TotalPrice - float64(cart.OfferPrice)
	order := &entity.Order{
		UserID:        cart.UserId,
		AddressId:     userAddress.ID,
		Total:         Total,
		Status:        "pending",
		PaymentMethod: "Cod",
		PaymentStatus: "pending",
	}

	OrderID, err2 := repo.Create(order)
	if err2 != nil {
		return nil, errors.New("Order placing failed")
	}
	invoiceData := &entity.Invoice{
		OrderId:     OrderID,
		UserId:      userId,
		AddressType: userAddress.Type,
		Quantity:    cart.TicketQuantity + cart.ApparelQuantity,
		Price:       order.Total,
		Payment:     order.PaymentMethod,
		Status:      order.PaymentStatus,
		PaymentId:   "nil",
		Remark:      "Zog_Festiv",
	}
	invoice, err := repo.CreateInvoice(invoiceData)
	if err != nil {
		return nil, errors.New("Invoice Creating failed")
	}
	for _, cartItem := range cartItems {
		orderItem := entity.OrderItem{
			OrderID:   OrderID,
			ProductID: cartItem.ProductId,
			Category:  cartItem.Category,
			Quantity:  cartItem.Quantity,
			Price:     cartItem.Price,
		}
		orderItems = append(orderItems, orderItem)
		inventory := entity.Inventory{
			ProductId:       cartItem.ProductId,
			ProductCategory: cartItem.Category,
			Quantity:        cartItem.Quantity,
		}
		err = repo.DecreaseProductQuantity(&inventory)
	}

	err = repo.CreateOrderItems(orderItems)
	if err != nil {
		return nil, errors.New("User cart is empty")
	}

	err = repo.RemoveCartItems(int(cart.ID))
	if err != nil {
		return nil, errors.New("Delete cart items failed")
	}
	cart.ApparelQuantity = 0
	cart.TicketQuantity = 0
	cart.TotalPrice = 0
	cart.OfferPrice = 0
	err = repo.UpdateCart(cart)
	if err != nil {
		return nil, errors.New("Updating cart failed")
	}
	return invoice, nil
}

func (s *Order) PaymentVerification(ctx context.Context, req *pb.PaymentVerificationRequest) (*pb.PaymentVerificationResponse, error) {
	result, err := repo.GetByRazorId(razorId)
	if err != nil {
		return nil, errors.New("Order not found")
	}
	err1 := utils.RazorPaymentVerification(Signature, razorId, paymentId)
	if err1 != nil {
		result.PaymentStatus = "failed"
		err2 := repo.Update(result)
		if err2 != nil {
			return nil, errors.New("payment updation failed")
		}
		return nil, err1
	}
	result.PaymentStatus = "successful"
	err3 := ou.orderRepo.Update(result)
	if err3 != nil {
		return nil, errors.New("payment updation failed")
	}
	userCart, err := repo.GetByUserID(result.UserID)
	if err != nil {
		return nil, errors.New("User cart not found")
	}
	userAddress, err := repo.GetAddressById(result.AddressId)
	if err != nil {
		return nil, errors.New("User address  not found")
	}
	Total := userCart.TotalPrice - float64(userCart.OfferPrice)
	invoiceData := &entity.Invoice{
		OrderId:     result.ID,
		UserId:      result.UserID,
		AddressType: userAddress.Type,
		Quantity:    userCart.TicketQuantity + userCart.ApparelQuantity,
		Price:       Total,
		Payment:     "razorpay",
		Status:      "succesful",
		PaymentId:   "nil",
		Remark:      "Zog_Festiv",
	}
	invoice, err := repo.CreateInvoice(invoiceData)
	if err != nil {
		return nil, errors.New("Invoice Creating failed")
	}
	err4 := repo.RemoveCartItems(int(userCart.ID))
	if err4 != nil {
		return nil, errors.New("Delete cart items failed")
	}
	userCart = &entity.Cart{
		OfferPrice:      0,
		TotalPrice:      0,
		TicketQuantity:  0,
		ApparelQuantity: 0,
	}
	err5 := repo.UpdateCart(userCart)
	if err5 != nil {
		return nil, errors.New("Updating cart failed")
	}
	return invoice, nil
}

func (s *Order) CancelOrder(ctx context.Context, req *pb.CancelOrderRequest) (*pb.CancelOrderResponse, error) {

}

func (s *Order) OrderHistory(ctx context.Context, req *pb.OrderHistoryRequest) (*pb.OrderHistoryResponse, error) {
	offset := (req.Page - 1) * req.Limit
	orderList, err := repo.GetAllOrders(int(req.Userid), int(offset), int(req.Limit))
	if err != nil {
		return nil, err
	}
	var pbOrderList []*pb.Orders
	for _, order := range orderList {
		pbOrder := &pb.Orders{
			ID:            int32(order.ID),
			UserID:        int32(order.UserID),
			AddressId:     int32(order.AddressId),
			Total:         int32(order.Total),
			Status:        order.Status,
			PaymentMethod: order.PaymentMethod,
			PaymentStatus: order.PaymentStatus,
			PaymentId:     order.PaymentId,
		}
		pbOrderList = append(pbOrderList, pbOrder)
	}

	response := &pb.OrderHistoryResponse{
		Order: pbOrderList,
	}

	return response, nil
}

func (s *Order) OrderReturn(ctx context.Context, req *pb.OrderReturnRequest) (*pb.OrderReturnResponse, error) {
	order, err := repo.GetByID(returnData.OrderId)
	if err != nil {
		return nil, errors.New("Order not found")
	}
	order.Status = "return"
	order.PaymentStatus = "refund"
	err = repo.Update(order)
	if err != nil {
		return nil, errors.New("order updation failed")
	}
	err = repo.CreateReturn(&returnData)
	if err != nil {
		return nil, errors.New("return creation failed")
	}
	return resp, nil
}
func (s *Order) AdminOrderUpdate(ctx context.Context, req *pb.AdminOrderUpdateRequest) (*pb.AdminOrderUpdateResponse, error) {
	result, err := repo.GetByID(orderId)
	if err != nil {
		return nil, errors.New("Order not found")
	}
	result.Status = status
	err = repo.Update(result)
	if err != nil {
		return nil, errors.New("order updation failed")
	}
	return resp, nil
}
func (s *Order) AdminReturnUpdate(ctx context.Context, req *pb.AdminReturnUpdateRequest) (*pb.AdminReturnUpdateResponse, error) {
	result, err := repo.GetReturnByID(returnId)
	if err != nil {
		return nil, errors.New("Order not found")
	}
	order, err := repo.GetByID(result.OrderId)
	if err != nil {
		return nil, errors.New("Order not found")
	}
	result.Status = status
	result.Refund = "wallet"
	result.TotalPrice = int(order.Total)
	err = repo.UpdateReturn(result)
	if err != nil {
		return nil, errors.New("return updation failed")
	}
	order.Status = "return"
	err = repo.Update(order)
	if err != nil {
		return nil, errors.New("order updation failed")
	}
	return resp, nil
}
func (s *Order) AdminRefund(ctx context.Context, req *pb.AdminRefundRequest) (*pb.AdminRefundResponse, error) {
	order, err := repo.GetByID(orderId)
	if err != nil {
		return nil, errors.New("Order not found")
	}

	if order.Status == "return" {
		result, err := repo.GetReturnByOrderID(order.ID)
		if err != nil {
			return nil, errors.New("Return not found")
		}
		user, err := repo.GetByID(result.UserId)
		if err != nil {
			return nil, errors.New("Order not found")
		}
		if result.Refund == "wallet" {
			user.Wallet = result.TotalPrice
			err = repo.UpdateUserWallet(user)
			if err != nil {
				return nil, errors.New("User wallet updation failed")
			}
		}
		result.Status = "completed"
		err = repo.UpdateReturn(result)
		if err != nil {
			return nil, errors.New("return updation failed")
		} else {
			return resp, nil
		}
	} else {
		user, err := repo.GetByID(order.UserID)
		if err != nil {
			return nil, errors.New("Order not found")
		}
		if order.PaymentStatus == "successful" {
			order.PaymentStatus = "refund"
			user.Wallet = int(order.Total)
			err = repo.UpdateUserWallet(user)
			if err != nil {
				return nil, errors.New("User wallet updation failed")
			}
		}
		order.Status = "canceled"
		err1 := repo.Update(order)
		if err1 != nil {
			return nil, errors.New("order cancelation failed")
		}
		return resp, nil
	}
}
func (s *Order) SalesReportByDate(ctx context.Context, req *pb.SalesReportByDateRequest) (*pb.SalesReportByDateResponse, error) {
	orders, err := repo.GetByDate(startDate, endDate)
	if err != nil {
		return nil, errors.New("report fetching failed")
	}
	return orders, nil
}
func (s *Order) SalesReportByPeriod(ctx context.Context, req *pb.SalesReportByPeriodRequest) (*pb.SalesReportByPeriodResponse, error) {
	startDate, endDate := utils.CalculatePeriodDates(period)

	orders, err := repo.GetByDate(startDate, endDate)
	if err != nil {
		return nil, errors.New("report fetching failed")
	}
	return orders, nil
}
func (s *Order) SalesReportByCategory(ctx context.Context, req *pb.SalesReportByCategoryRequest) (*pb.SalesReportByCategoryResponse, error) {
	startDate, endDate := utils.CalculatePeriodDates(period)
	orders, err := repo.GetByCategory(category, startDate, endDate)
	if err != nil {
		return nil, errors.New("report fetching failed")
	}
	return orders, nil
}
func (s *Order) SortOrderByStatus(ctx context.Context, req *pb.SortOrderByStatusRequest) (*pb.SortOrderByStatusResponse, error) {
	offset := (page - 1) * limit
	orders, err := repo.GetByStatus(offset, limit, status)
	if err != nil {
		return nil, errors.New("report fetching failed")
	}
	return orders, nil
}
