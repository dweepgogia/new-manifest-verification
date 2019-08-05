package validator

import "fmt"

// ManifestResult represents verification result for each of the yaml files
// from the manifest bundle.
type ManifestResult struct {
	// Name is some piece of information identifying the manifest. This should
	// usually be set to object.GetName().
	Name string
	// Errors pertain to issues with the manifest that must be corrected.
	Errors []Error
	// Warnings pertain to issues with the manifest that are optional to correct.
	Warnings []Error
}

// Error is an implementation of the 'error' interface, which represents a
// warning or an error in a yaml file. Error type is taken as is from
// https://github.com/operator-framework/operator-registry/blob/master/vendor/k8s.io/apimachinery/pkg/util/validation/field/errors.go#L31
// to maintain compatibility with upstream.
type Error struct {
	// Type is the ErrorType string constant that represents the kind of
	// error, ex. "MandatoryStructMissing", "I/O".
	Type ErrorType
	// Field is the dot-hierarchical YAML path of the missing data.
	Field string
	// BadValue is the field or file that caused an error or warning.
	BadValue interface{}
	// Detail represents the error message as a string.
	Detail string
}

func (err Error) String() string {
	//return fmt.Sprintf("Detail: %s | Error type: %s | Value: %v | Field: %s", err.Detail, err.Type.String(), err.BadValue, err.Field)
	return err.Error()
}

type ErrorType string

func InvalidCSV(detail string) Error {
	return Error{ErrorInvalidCSV, "", "", detail}
}

func InvalidDefaultChannel(detail string, value interface{}) Error {
	return Error{ErrorInvalidDefaultChannel, "", value, detail}
}

func OptionalFieldMissing(detail string, field string, value interface{}) Error {
	return Error{WarningFieldMissing, field, value, detail}
}

func MandatoryFieldMissing(detail string, field string, value interface{}) Error {
	return Error{ErrorFieldMissing, field, value, detail}
}

func UnsupportedType(detail string) Error {
	return Error{ErrorUnsupportedType, "", "", detail}
}

// TODO: see if more information can be extracted out of 'unmarshall/parsing' errors.
func InvalidParse(detail string, value interface{}) Error {
	return Error{ErrorInvalidParse, "", value, detail}
}

func IOError(detail string, value interface{}) Error {
	return Error{ErrorIO, "", value, detail}
}

func FailedValidation(detail string, value interface{}) Error {
	return Error{ErrorFailedValidation, "", value, detail}
}

func InvalidOperation(detail string, value interface{}) Error {
	return Error{ErrorInvalidOperation, "", value, detail}
}

const (
	ErrorInvalidCSV            ErrorType = "CSVFileNotValid"
	WarningFieldMissing        ErrorType = "OptionalFieldNotFound"
	ErrorFieldMissing          ErrorType = "MandatoryFieldNotFound"
	ErrorUnsupportedType       ErrorType = "FieldTypeNotSupported"
	ErrorInvalidParse          ErrorType = "Unmarshall/ParseError"
	ErrorIO                    ErrorType = "FileReadError"
	ErrorFailedValidation      ErrorType = "ValidationFailed"
	ErrorInvalidOperation      ErrorType = "OperationFailed"
	ErrorInvalidDefaultChannel ErrorType = "DefaultChannelNotValid"
)

// String converts a ErrorType into its corresponding canonical error message.
func (t ErrorType) String() string {
	switch t {
	case ErrorInvalidCSV:
		return "CSV file not valid"
	case WarningFieldMissing:
		return "Optional field not found"
	case ErrorFieldMissing:
		return "Mandatory field not found"
	case ErrorUnsupportedType:
		return "Field type not supported"
	case ErrorInvalidParse:
		return "Unmarshall/Parse error"
	case ErrorIO:
		return "File read error"
	case ErrorFailedValidation:
		return "Validation failed"
	case ErrorInvalidOperation:
		return "Operation failed"
	case ErrorInvalidDefaultChannel:
		return "Default channel not found"
	default:
		panic(fmt.Sprintf("Unrecognized validation error: %q", string(t)))
	}
}

// Error strut implements the 'error' interface to define custom error formatting.
func (err Error) Error() string {
	return err.Detail
}

// ValidatorSet contains a set of Validators to be executed sequentially.
// TODO: add configurable logger.
type ValidatorSet struct {
	validators []Validator
}

// NewValidatorSet creates a ValidatorSet containing vs.
func NewValidatorSet(vs ...Validator) *ValidatorSet {
	set := &ValidatorSet{}
	set.AddValidators(vs...)
	return set
}

// AddValidators adds each unique Validator in vs to the receiver.
func (set *ValidatorSet) AddValidators(vs ...Validator) {
	seenNames := map[string]struct{}{}
	for _, v := range vs {
		if _, seen := seenNames[v.Name()]; !seen {
			set.validators = append(set.validators, v)
			seenNames[v.Name()] = struct{}{}
		}
	}
}

// ValidateAll runs each Validator in the receiver and returns all results.
func (set ValidatorSet) ValidateAll() (allResults []ManifestResult) {
	for _, v := range set.validators {
		results := v.Validate()
		allResults = append(allResults, results...)
	}
	return allResults
}
