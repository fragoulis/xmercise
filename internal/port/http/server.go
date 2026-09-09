package http

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/oapi-codegen/nullable"
	"github.com/samber/lo"

	appcompany "github.com/fragoulis/xmercise/internal/application/company"
	domaincompany "github.com/fragoulis/xmercise/internal/domain/company"
)

// Server implements the generated strict HTTP port.
type Server struct {
	companies *appcompany.Service
}

// NewServer creates an HTTP server adapter.
func NewServer(companies *appcompany.Service) *Server {
	return &Server{
		companies: companies,
	}
}

// CreateCompany creates a company.
func (s *Server) CreateCompany(
	ctx context.Context,
	request CreateCompanyRequestObject,
) (CreateCompanyResponseObject, error) {
	created, err := s.companies.Create(ctx, appcompany.CreateCommand{
		Name:           request.Body.Name,
		Description:    nullableValue(request.Body.Description),
		EmployeesCount: int(request.Body.EmployeesCount),
		Registered:     request.Body.Registered,
		Type:           request.Body.Type,
	})
	if err != nil {
		return createErrorResponse(err)
	}

	return CreateCompany201JSONResponse(companyResponse(created)), nil
}

// UpdateCompany updates a company.
func (s *Server) UpdateCompany(
	ctx context.Context,
	request UpdateCompanyRequestObject,
) (UpdateCompanyResponseObject, error) {
	body := request.JSONBody
	if body == nil {
		body = request.ApplicationMergePatchPlusJSONBody
	}
	if body == nil {
		return updateBadRequest("request body is required"), nil
	}

	command := appcompany.UpdateCommand{
		ID:   request.Id,
		Name: body.Name,
		Description: domaincompany.DescriptionPatch{
			Present: body.Description.IsSpecified(),
			Value:   nullableValue(body.Description),
		},
		Registered: body.Registered,
		Type:       body.Type,
	}
	if body.EmployeesCount != nil {
		command.EmployeesCount = lo.ToPtr(int(*body.EmployeesCount))
	}

	updated, err := s.companies.Update(ctx, command)
	if err != nil {
		return updateErrorResponse(err)
	}

	return UpdateCompany200JSONResponse(companyResponse(updated)), nil
}

// DeleteCompany deletes a company.
func (s *Server) DeleteCompany(
	ctx context.Context,
	request DeleteCompanyRequestObject,
) (DeleteCompanyResponseObject, error) {
	err := s.companies.Delete(ctx, appcompany.DeleteCommand{
		ID: request.Id,
	})
	if validation, ok := validationError(err); ok {
		return DeleteCompany400JSONResponse{
			BadRequestJSONResponse: BadRequestJSONResponse(validation),
		}, nil
	}
	if errors.Is(err, appcompany.ErrCompanyNotFound) {
		return DeleteCompany404JSONResponse{
			NotFoundJSONResponse: NotFoundJSONResponse(errorResponse(404, "Company not found", "")),
		}, nil
	}
	if err != nil {
		return nil, err
	}

	return DeleteCompany204Response{}, nil
}

// GetCompany gets a company.
func (s *Server) GetCompany(
	ctx context.Context,
	request GetCompanyRequestObject,
) (GetCompanyResponseObject, error) {
	found, err := s.companies.FindOne(ctx, appcompany.FindOneQuery{
		ID: request.Id,
	})
	if validation, ok := validationError(err); ok {
		return GetCompany400JSONResponse{
			BadRequestJSONResponse: BadRequestJSONResponse(validation),
		}, nil
	}
	if errors.Is(err, appcompany.ErrCompanyNotFound) {
		return GetCompany404JSONResponse{
			NotFoundJSONResponse: NotFoundJSONResponse(errorResponse(404, "Company not found", "")),
		}, nil
	}
	if err != nil {
		return nil, err
	}

	return GetCompany200JSONResponse(companyResponse(found)), nil
}

func nullableValue(value nullable.Nullable[string]) *string {
	if !value.IsSpecified() || value.IsNull() {
		return nil
	}

	return lo.ToPtr(value.MustGet())
}

func companyResponse(company *domaincompany.Company) Company {
	description := nullable.NewNullNullable[string]()
	if value := company.Description(); value != nil {
		description = nullable.NewNullableWithValue(*value)
	}

	id := company.ID().String()
	createdAt := company.CreatedAt()
	updatedAt := company.UpdatedAt()
	return Company{
		Id:             &id,
		Name:           company.Name(),
		Description:    description,
		EmployeesCount: int32(company.EmployeesCount()),
		Registered:     company.Registered(),
		Type:           string(company.Type()),
		CreatedAt:      &createdAt,
		UpdatedAt:      &updatedAt,
	}
}

func createErrorResponse(err error) (CreateCompanyResponseObject, error) {
	if validation, ok := validationError(err); ok {
		return CreateCompany400JSONResponse{
			BadRequestJSONResponse: BadRequestJSONResponse(validation),
		}, nil
	}
	if errors.Is(err, appcompany.ErrCompanyNameTaken) {
		return CreateCompany409JSONResponse{
			ConflictJSONResponse: ConflictJSONResponse(errorResponse(409, "Company name is already taken", "")),
		}, nil
	}

	return nil, err
}

func updateErrorResponse(err error) (UpdateCompanyResponseObject, error) {
	if validation, ok := validationError(err); ok {
		return UpdateCompany400JSONResponse{
			BadRequestJSONResponse: BadRequestJSONResponse(validation),
		}, nil
	}
	if errors.Is(err, appcompany.ErrCompanyNotFound) {
		return UpdateCompany404JSONResponse{
			NotFoundJSONResponse: NotFoundJSONResponse(errorResponse(404, "Company not found", "")),
		}, nil
	}
	if errors.Is(err, appcompany.ErrCompanyNameTaken) {
		return UpdateCompany409JSONResponse{
			ConflictJSONResponse: ConflictJSONResponse(errorResponse(409, "Company name is already taken", "")),
		}, nil
	}

	return nil, err
}

func updateBadRequest(detail string) UpdateCompany400JSONResponse {
	return UpdateCompany400JSONResponse{
		BadRequestJSONResponse: BadRequestJSONResponse(errorResponse(400, "Invalid request", detail)),
	}
}

func validationError(err error) (Error, bool) {
	var validation appcompany.ValidationError
	if !errors.As(err, &validation) {
		return Error{}, false
	}

	violations := lo.Map(validation.Violations, func(violation appcompany.Violation, _ int) Violation {
		return Violation{
			Field:   violation.Field,
			Message: violation.Message,
		}
	})
	response := errorResponse(400, "Invalid request", "")
	response.Violations = &violations
	return response, true
}

// NewJSONStrictHandler creates a strict handler with JSON error responses.
func NewJSONStrictHandler(server StrictServerInterface) ServerInterface {
	return NewStrictHandlerWithOptions(server, nil, StrictHTTPServerOptions{
		RequestErrorHandlerFunc: func(w http.ResponseWriter, _ *http.Request, _ error) {
			writeError(w, http.StatusBadRequest, "Invalid request", "Invalid request body")
		},
		ResponseErrorHandlerFunc: func(w http.ResponseWriter, _ *http.Request, _ error) {
			writeError(w, http.StatusInternalServerError, "Internal server error", "")
		},
	})
}

// NewJSONHandler creates a router with JSON error responses.
func NewJSONHandler(server ServerInterface) http.Handler {
	router := chi.NewRouter()
	router.NotFound(func(w http.ResponseWriter, _ *http.Request) {
		writeError(w, http.StatusNotFound, "Not found", "")
	})
	router.MethodNotAllowed(func(w http.ResponseWriter, _ *http.Request) {
		writeError(w, http.StatusMethodNotAllowed, "Method not allowed", "")
	})

	return HandlerWithOptions(server, ChiServerOptions{
		BaseRouter: router,
		ErrorHandlerFunc: func(w http.ResponseWriter, _ *http.Request, _ error) {
			writeError(w, http.StatusBadRequest, "Invalid request", "Invalid request parameter")
		},
	})
}

func writeError(w http.ResponseWriter, status int, title, detail string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(errorResponse(int32(status), title, detail))
}

func errorResponse(status int32, title, detail string) Error {
	response := Error{
		Status: status,
		Title:  title,
	}
	if detail != "" {
		response.Detail = &detail
	}

	return response
}
