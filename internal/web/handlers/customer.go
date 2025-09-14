package handlers

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/cr1cr1/farm-manager/internal/data"
	"github.com/cr1cr1/farm-manager/internal/domain"
	"github.com/cr1cr1/farm-manager/internal/web/middleware"
	"github.com/cr1cr1/farm-manager/internal/web/templates/pages"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"
)

type CustomerManager struct {
	CustomerRepo data.CustomerRepo
}

// RegisterCustomerRoutes wires customer management endpoints under /app.
func RegisterCustomerRoutes(group *ghttp.RouterGroup, customerRepo data.CustomerRepo) {
	cm := &CustomerManager{
		CustomerRepo: customerRepo,
	}

	// Customer management
	group.GET("/management/customers", cm.CustomersGet)
	group.POST("/management/customers", cm.CustomerPost)
	group.GET("/management/customers/new", cm.CustomerGet)
	group.GET("/management/customers/:id", cm.CustomerGet)
	group.PUT("/management/customers/:id", cm.CustomerPut)
	group.DELETE("/management/customers/:id", cm.CustomerDelete)
}

// CustomersGet renders the customers management page.
func (cm *CustomerManager) CustomersGet(r *ghttp.Request) {
	user, ok := middleware.CurrentUser(r)
	if !ok {
		r.Response.RedirectTo(middleware.BasePath() + "/login")
		return
	}

	customers, err := cm.CustomerRepo.List(r.GetCtx())
	if err != nil {
		g.Log().Errorf(r.GetCtx(), "list customers: %v", err)
		r.Response.WriteStatusExit(500, "Internal server error")
		return
	}

	isDataStarRequest := r.Header.Get("datastar-request") == "true"
	if isDataStarRequest {
		_ = middleware.TemplRender(
			r,
			pages.CustomersContent(
				middleware.BasePath(),
				middleware.CsrfToken(r),
				customers,
			),
		)
		return
	}

	_ = middleware.TemplRender(
		r,
		pages.CustomersPage(
			middleware.BasePath(),
			middleware.CsrfToken(r),
			user.Username,
			ThemeToString(user.Theme),
			customers,
		),
	)
}

// CustomerPost creates a new customer.
func (cm *CustomerManager) CustomerPost(r *ghttp.Request) {
	user, ok := middleware.CurrentUser(r)
	if !ok {
		r.Response.RedirectTo(middleware.BasePath() + "/login")
		return
	}

	name := strings.TrimSpace(r.Get("name").String())
	contactInfo := strings.TrimSpace(r.Get("contact_info").String())
	deliveryAddress := strings.TrimSpace(r.Get("delivery_address").String())
	customerType := strings.TrimSpace(r.Get("customer_type").String())

	errs := map[string]string{}
	if name == "" {
		errs["name"] = "Name is required"
	}

	var contact *string
	if contactInfo != "" {
		contact = new(string)
		*contact = contactInfo
	}

	var deliveryAddr *string
	if deliveryAddress != "" {
		deliveryAddr = new(string)
		*deliveryAddr = deliveryAddress
	}

	var custType *string
	if customerType != "" {
		custType = new(string)
		*custType = customerType
	}

	isDataStarRequest := r.Header.Get("datastar-request") == "true"

	if len(errs) == 0 {
		userIDStr := strconv.FormatInt(user.ID, 10)
		createdBy := new(string)
		*createdBy = userIDStr
		updatedBy := new(string)
		*updatedBy = userIDStr
		customer := &domain.Customer{
			Name:            name,
			ContactInfo:     contact,
			DeliveryAddress: deliveryAddr,
			CustomerType:    custType,
			Audit: domain.AuditFields{
				CreatedBy: createdBy,
				UpdatedBy: updatedBy,
			},
		}

		_, err := cm.CustomerRepo.Create(r.GetCtx(), customer)
		if err != nil {
			g.Log().Errorf(r.GetCtx(), "create customer: %v", err)
			errs["form"] = "Failed to create customer"
		}
	}

	if len(errs) > 0 {
		if isDataStarRequest {
			// For DataStar requests, return validation errors
			r.Response.Header().Set("Content-Type", "application/json")
			r.Response.WriteJson(map[string]interface{}{
				"errors": errs,
			})
			return
		}
		// For regular requests, redirect back with errors
		r.Response.RedirectTo(middleware.BasePath() + "/management/customers")
		return
	}

	if isDataStarRequest {
		// For DataStar requests, redirect via JavaScript
		js := fmt.Sprintf("window.location.href = %q;", middleware.BasePath()+"/management/customers")
		r.Response.Header().Set("Content-Type", "text/javascript")
		r.Response.Write([]byte(js))
		return
	}

	// For regular requests, redirect to the list
	r.Response.RedirectTo(middleware.BasePath() + "/management/customers")
}

// CustomerGet renders a specific customer for editing or a new customer form.
func (cm *CustomerManager) CustomerGet(r *ghttp.Request) {
	user, ok := middleware.CurrentUser(r)
	if !ok {
		r.Response.RedirectTo(middleware.BasePath() + "/login")
		return
	}

	idStr := r.Get("id").String()
	var customer *domain.Customer

	// Check if this is a request for a new customer (no ID provided)
	if idStr == "" || idStr == "new" {
		// This is a request for creating a new customer
		customer = nil
	} else {
		// This is a request for editing an existing customer
		id, err := strconv.ParseInt(idStr, 10, 64)
		if err != nil {
			r.Response.WriteStatusExit(400, "Invalid customer ID")
			return
		}

		customer, err = cm.CustomerRepo.FindByID(r.GetCtx(), id)
		if err != nil {
			if err == data.ErrNotFound {
				r.Response.WriteStatusExit(404, "Customer not found")
				return
			}
			g.Log().Errorf(r.GetCtx(), "find customer: %v", err)
			r.Response.WriteStatusExit(500, "Internal server error")
			return
		}
	}

	isDataStarRequest := r.Header.Get("datastar-request") == "true"
	if isDataStarRequest {
		_ = middleware.TemplRender(
			r,
			pages.CustomerContent(
				middleware.BasePath(),
				middleware.CsrfToken(r),
				customer,
			),
		)
		return
	}

	_ = middleware.TemplRender(
		r,
		pages.CustomerPage(
			middleware.BasePath(),
			middleware.CsrfToken(r),
			user.Username,
			ThemeToString(user.Theme),
			customer,
		),
	)
}

// CustomerPut updates an existing customer.
func (cm *CustomerManager) CustomerPut(r *ghttp.Request) {
	user, ok := middleware.CurrentUser(r)
	if !ok {
		r.Response.RedirectTo(middleware.BasePath() + "/login")
		return
	}

	idStr := r.Get("id").String()
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		r.Response.WriteStatusExit(400, "Invalid customer ID")
		return
	}

	name := strings.TrimSpace(r.Get("name").String())
	contactInfo := strings.TrimSpace(r.Get("contact_info").String())
	deliveryAddress := strings.TrimSpace(r.Get("delivery_address").String())
	customerType := strings.TrimSpace(r.Get("customer_type").String())

	errs := map[string]string{}
	if name == "" {
		errs["name"] = "Name is required"
	}

	var contact *string
	if contactInfo != "" {
		contact = new(string)
		*contact = contactInfo
	}

	var deliveryAddr *string
	if deliveryAddress != "" {
		deliveryAddr = new(string)
		*deliveryAddr = deliveryAddress
	}

	var custType *string
	if customerType != "" {
		custType = new(string)
		*custType = customerType
	}

	isDataStarRequest := r.Header.Get("datastar-request") == "true"

	if len(errs) == 0 {
		userIDStr := strconv.FormatInt(user.ID, 10)
		updatedBy := new(string)
		*updatedBy = userIDStr
		customer := &domain.Customer{
			CustomerID:      id,
			Name:            name,
			ContactInfo:     contact,
			DeliveryAddress: deliveryAddr,
			CustomerType:    custType,
			Audit: domain.AuditFields{
				UpdatedBy: updatedBy,
			},
		}

		err := cm.CustomerRepo.Update(r.GetCtx(), customer)
		if err != nil {
			g.Log().Errorf(r.GetCtx(), "update customer: %v", err)
			errs["form"] = "Failed to update customer"
		}
	}

	if len(errs) > 0 {
		if isDataStarRequest {
			// For DataStar requests, return validation errors
			r.Response.Header().Set("Content-Type", "application/json")
			r.Response.WriteJson(map[string]interface{}{
				"errors": errs,
			})
			return
		}
		// For regular requests, redirect back with errors
		r.Response.RedirectTo(fmt.Sprintf("%s/management/customers/%d", middleware.BasePath(), id))
		return
	}

	if isDataStarRequest {
		// For DataStar requests, redirect via JavaScript
		js := fmt.Sprintf("window.location.href = %q;", fmt.Sprintf("%s/management/customers/%d", middleware.BasePath(), id))
		r.Response.Header().Set("Content-Type", "text/javascript")
		r.Response.Write([]byte(js))
		return
	}

	// For regular requests, redirect to the updated customer
	r.Response.RedirectTo(fmt.Sprintf("%s/management/customers/%d", middleware.BasePath(), id))
}

// CustomerDelete soft deletes a customer.
func (cm *CustomerManager) CustomerDelete(r *ghttp.Request) {
	_, ok := middleware.CurrentUser(r)
	if !ok {
		r.Response.RedirectTo(middleware.BasePath() + "/login")
		return
	}

	idStr := r.Get("id").String()
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		r.Response.WriteStatusExit(400, "Invalid customer ID")
		return
	}

	err = cm.CustomerRepo.SoftDelete(r.GetCtx(), id, time.Now())
	if err != nil {
		g.Log().Errorf(r.GetCtx(), "delete customer: %v", err)
		r.Response.WriteStatusExit(500, "Internal server error")
		return
	}

	isDataStarRequest := r.Header.Get("datastar-request") == "true"
	if isDataStarRequest {
		// For DataStar requests, redirect via JavaScript
		js := fmt.Sprintf("window.location.href = %q;", middleware.BasePath()+"/management/customers")
		r.Response.Header().Set("Content-Type", "text/javascript")
		r.Response.Write([]byte(js))
		return
	}

	// For regular requests, redirect to the list
	r.Response.RedirectTo(middleware.BasePath() + "/management/customers")
}
