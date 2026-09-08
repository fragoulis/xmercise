package company

import (
	"strings"
	"unicode/utf8"

	"github.com/google/uuid"

	companydomain "github.com/fragoulis/xmercise/internal/domain/company"
)

const (
	maxNameLength        = 15
	maxDescriptionLength = 3000
)

func validateCreateCommand(command CreateCommand) error {
	return validationError(validateCompanyFields(
		&command.Name,
		command.Description,
		&command.EmployeesCount,
		&command.Type,
	))
}

func validateUpdateCommand(command UpdateCommand) error {
	var violations []Violation
	if command.ID == uuid.Nil {
		violations = append(violations, Violation{
			Field:   "id",
			Message: "is required",
		})
	}

	var description *string
	if command.Description.Present {
		description = command.Description.Value
	}

	return validationError(append(violations, validateCompanyFields(
		command.Name,
		description,
		command.EmployeesCount,
		command.Type,
	)...))
}

func validateID(id uuid.UUID) error {
	if id == uuid.Nil {
		return ValidationError{
			Violations: []Violation{
				{
					Field:   "id",
					Message: "is required",
				},
			},
		}
	}

	return nil
}

func validateCompanyFields(
	name *string,
	description *string,
	employeesCount *int,
	companyType *string,
) []Violation {
	var violations []Violation
	if name != nil {
		switch {
		case strings.TrimSpace(*name) == "":
			violations = append(violations, Violation{
				Field:   "name",
				Message: "is required",
			})
		case utf8.RuneCountInString(*name) > maxNameLength:
			violations = append(violations, Violation{
				Field:   "name",
				Message: "must be at most 15 characters",
			})
		}
	}
	if description != nil && utf8.RuneCountInString(*description) > maxDescriptionLength {
		violations = append(violations, Violation{
			Field:   "description",
			Message: "must be at most 3000 characters",
		})
	}
	if employeesCount != nil && *employeesCount < 0 {
		violations = append(violations, Violation{
			Field:   "employees_count",
			Message: "must be greater than or equal to 0",
		})
	}
	if companyType != nil && !validCompanyType(*companyType) {
		violations = append(violations, Violation{
			Field:   "type",
			Message: "is invalid",
		})
	}

	return violations
}

func validCompanyType(value string) bool {
	switch companydomain.Type(value) {
	case companydomain.TypeCorporations,
		companydomain.TypeNonProfit,
		companydomain.TypeCooperative,
		companydomain.TypeSoleProprietorship:
		return true
	default:
		return false
	}
}

func validationError(violations []Violation) error {
	if len(violations) == 0 {
		return nil
	}

	return ValidationError{
		Violations: violations,
	}
}
