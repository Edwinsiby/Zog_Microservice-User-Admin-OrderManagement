package service

import (
	"context"
	"errors"
	"fmt"
	"log"
	pb "service4/pb"
	"service4/pkg/entity"
	repo "service4/pkg/repository"
	utils "service4/pkg/utils"

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
func ExecutePurchaseRazorPay(userId int, address int) (string, int, error) {
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
	OrderId, err2 := repo.CreateOrder(order)
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

	OrderID, err2 := repo.CreateOrder(order)
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
	result, err := repo.GetByRazorId(req.Razorid)
	if err != nil {
		return nil, errors.New("Order not found")
	}
	err = utils.RazorPaymentVerification(req.Signature, req.Razorid, req.Paymentid)
	if err != nil {
		result.PaymentStatus = "failed"
		err = repo.Update(result)
		if err != nil {
			return nil, errors.New("payment updation failed")
		}
		return nil, err
	}
	result.PaymentStatus = "successful"
	err = repo.Update(result)
	if err != nil {
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
	err = repo.RemoveCartItems(int(userCart.ID))
	if err != nil {
		return nil, errors.New("Delete cart items failed")
	}
	userCart = &entity.Cart{
		OfferPrice:      0,
		TotalPrice:      0,
		TicketQuantity:  0,
		ApparelQuantity: 0,
	}
	err = repo.UpdateCart(userCart)
	if err != nil {
		return nil, errors.New("Updating cart failed")
	}

	response := &pb.PaymentVerificationResponse{
		Result:    "Payment Successful",
		Paymentid: invoice.PaymentId,
	}
	return response, nil
}

func (s *Order) CancelOrder(ctx context.Context, req *pb.CancelOrderRequest) (*pb.CancelOrderResponse, error) {
	result, err := repo.GetByID(int(req.Orderid))
	if err != nil {
		return nil, errors.New("Order not found")
	}
	user, err := repo.GetUserByID(result.UserID)
	if err != nil {
		return nil, errors.New("User not found")
	}
	if result.Status != "pending" && result.Status != "confirmed" {
		return nil, errors.New("order cancelation failed- cancel time exceeded")
	}
	if result.PaymentStatus == "successful" {
		result.PaymentStatus = "refund"
		user.Wallet = int(result.Total)
		err = repo.UpdateUserWallet(user)
		if err != nil {
			return nil, errors.New("User wallet updation failed")
		}
	}
	result.Status = "canceled"
	err1 := repo.Update(result)
	if err1 != nil {
		return nil, errors.New("order cancelation failed")
	}
	resp := &pb.CancelOrderResponse{Result: "Order Canceled"}
	return resp, nil
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
	order, err := repo.GetByID(int(req.Orderid))
	if err != nil {
		return nil, errors.New("Order not found")
	}
	order.Status = "return"
	order.PaymentStatus = "refund"
	err = repo.Update(order)
	if err != nil {
		return nil, errors.New("order updation failed")
	}
	returnData := entity.Return{
		UserId:  int(req.Userid),
		OrderId: int(req.Orderid),
		Reason:  req.Reason,
		Status:  req.Status,
	}
	err = repo.CreateReturn(&returnData)
	if err != nil {
		return nil, errors.New("return creation failed")
	}
	resp := &pb.OrderReturnResponse{Result: "Order return requested"}
	return resp, nil
}
func (s *Order) AdminOrderUpdate(ctx context.Context, req *pb.AdminOrderUpdateRequest) (*pb.AdminOrderUpdateResponse, error) {
	result, err := repo.GetByID(int(req.Orderid))
	if err != nil {
		return nil, errors.New("Order not found")
	}
	result.Status = req.Status
	err = repo.Update(result)
	if err != nil {
		return nil, errors.New("order updation failed")
	}
	resp := &pb.AdminOrderUpdateResponse{Result: "Order Status Updated"}
	return resp, nil
}

func (s *Order) AdminReturnUpdate(ctx context.Context, req *pb.AdminReturnUpdateRequest) (*pb.AdminReturnUpdateResponse, error) {
	result, err := repo.GetReturnByID(int(req.Returnid))
	if err != nil {
		return nil, errors.New("Order not found")
	}
	order, err := repo.GetByID(result.OrderId)
	if err != nil {
		return nil, errors.New("Order not found")
	}
	result.Status = req.Status
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
	resp := &pb.AdminReturnUpdateResponse{Result: "Order Return Status Updated"}
	return resp, nil
}

func (s *Order) AdminRefund(ctx context.Context, req *pb.AdminRefundRequest) (*pb.AdminRefundResponse, error) {
	order, err := repo.GetByID(int(req.Orderid))
	if err != nil {
		return nil, errors.New("Order not found")
	}

	if order.Status == "return" {
		result, err := repo.GetReturnByOrderID(order.ID)
		if err != nil {
			return nil, errors.New("Return not found")
		}
		user, err := repo.GetUserByID(result.UserId)
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
			resp := &pb.AdminRefundResponse{Result: "Order Refund Status Updated"}
			return resp, nil
		}
	} else {
		user, err := repo.GetUserByID(order.UserID)
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
		resp := &pb.AdminRefundResponse{Result: "Order Refund Status Updated"}
		return resp, nil
	}
}

func (s *Order) SalesReportByDate(ctx context.Context, req *pb.SalesReportByDateRequest) (*pb.SalesReportByDateResponse, error) {
	fmt.Println("SalesReposrtBYDate")
	resp := &pb.SalesReportByDateResponse{}
	return resp, nil
}

func (s *Order) SalesReportByPeriod(ctx context.Context, req *pb.SalesReportByPeriodRequest) (*pb.SalesReportByPeriodResponse, error) {
	startDate, endDate := utils.CalculatePeriodDates(req.Period)

	orders, err := repo.GetByDate(startDate, endDate)
	if err != nil {
		return nil, errors.New("report fetching failed")
	}
	response := &pb.SalesReportByPeriodResponse{
		TotalSales:   int32(orders.TotalSales),
		TotalOrders:  int32(orders.TotalOrders),
		AverageOrder: int32(orders.AverageOrder),
	}
	return response, nil
}

func (s *Order) SalesReportByCategory(ctx context.Context, req *pb.SalesReportByCategoryRequest) (*pb.SalesReportByCategoryResponse, error) {
	startDate, endDate := utils.CalculatePeriodDates(req.Period)
	orders, err := repo.GetByCategory(req.Category, startDate, endDate)
	if err != nil {
		return nil, errors.New("report fetching failed")
	}
	response := &pb.SalesReportByCategoryResponse{
		TotalSales:   int32(orders.TotalSales),
		TotalOrders:  int32(orders.TotalOrders),
		AverageOrder: int32(orders.AverageOrder),
	}
	return response, nil
}

func (s *Order) SortOrderByStatus(ctx context.Context, req *pb.SortOrderByStatusRequest) (*pb.SortOrderByStatusResponse, error) {
	offset := (req.Page - 1) * req.Limit
	orders, err := repo.GetByStatus(int(offset), int(req.Limit), req.Status)
	if err != nil {
		return nil, errors.New("report fetching failed")
	}
	var pbOrderList []*pb.Orders
	for _, order := range orders {
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

	response := &pb.SortOrderByStatusResponse{
		Order: pbOrderList,
	}
	return response, nil
}
