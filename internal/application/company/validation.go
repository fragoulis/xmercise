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
	violations := requiredCreateViolations(command)
	violations = append(violations, validateCompanyFields(
		&command.Name,
		command.Description,
		&command.EmployeesCount,
		&command.Type,
	)...)

	return validationError(violations)
}

func requiredCreateViolations(command CreateCommand) []Violation {
	var violations []Violation
	if strings.TrimSpace(command.Name) == "" {
		violations = append(violations, Violation{
			Field:   "name",
			Message: "is required",
		})
	}
	if strings.TrimSpace(command.Type) == "" {
		violations = append(violations, Violation{
			Field:   "type",
			Message: "is required",
		})
	}

	return violations
}

func validateUpdateCommand(command UpdateCommand) (uuid.UUID, error) {
	id, idViolation := parseID(command.ID)
	violations := validateCompanyFields(
		command.Name,
		descriptionValue(command.Description),
		command.EmployeesCount,
		command.Type,
	)
	if idViolation != nil {
		violations = append([]Violation{*idViolation}, violations...)
	}

	return id, validationError(violations)
}

func validateID(value string) (uuid.UUID, error) {
	id, violation := parseID(value)
	if violation == nil {
		return id, nil
	}

	return uuid.Nil, ValidationError{
		Violations: []Violation{*violation},
	}
}

func parseID(value string) (uuid.UUID, *Violation) {
	if strings.TrimSpace(value) == "" {
		return uuid.Nil, &Violation{
			Field:   "id",
			Message: "is required",
		}
	}

	id, err := uuid.Parse(value)
	if err != nil {
		return uuid.Nil, &Violation{
			Field:   "id",
			Message: "must be a valid UUID",
		}
	}
	if id == uuid.Nil {
		return uuid.Nil, &Violation{
			Field:   "id",
			Message: "is required",
		}
	}

	return id, nil
}

func descriptionValue(patch companydomain.DescriptionPatch) *string {
	if !patch.Present {
		return nil
	}

	return patch.Value
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
