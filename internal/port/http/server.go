package http

import (
	"context"
	"errors"

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
	if request.Body == nil {
		return createBadRequest("request body is required"), nil
	}

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
		ID:          request.Id,
		Name:        body.Name,
		Description: descriptionPatch(body.Description),
		Registered:  body.Registered,
		Type:        body.Type,
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

func descriptionPatch(value nullable.Nullable[string]) domaincompany.DescriptionPatch {
	return domaincompany.DescriptionPatch{
		Present: value.IsSpecified(),
		Value:   nullableValue(value),
	}
}

func companyResponse(company *domaincompany.Company) Company {
	description := nullable.NewNullNullable[string]()
	if value, ok := company.Description(); ok {
		description = nullable.NewNullableWithValue(value)
	}

	id := company.ID()
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

func validationError(err error) (Error, bool) {
	var validation domaincompany.ValidationError
	if !errors.As(err, &validation) {
		return Error{}, false
	}

	violations := lo.Map(validation.Violations, func(violation domaincompany.Violation, _ int) Violation {
		return Violation{
			Field:   violation.Field,
			Message: violation.Message,
		}
	})
	response := errorResponse(400, "Invalid request", "")
	response.Violations = &violations
	return response, true
}

func createBadRequest(detail string) CreateCompany400JSONResponse {
	return CreateCompany400JSONResponse{
		BadRequestJSONResponse: BadRequestJSONResponse(errorResponse(400, "Invalid request", detail)),
	}
}

func updateBadRequest(detail string) UpdateCompany400JSONResponse {
	return UpdateCompany400JSONResponse{
		BadRequestJSONResponse: BadRequestJSONResponse(errorResponse(400, "Invalid request", detail)),
	}
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
