// Package domain provides core domain types for the exchange connector.
// All financial values use decimal arithmetic via cockroachdb/apd for precision.
package domain

import (
	"fmt"

	"github.com/cockroachdb/apd/v3"
)

// Decimal is a type alias for apd.Decimal pointer, providing ergonomic decimal arithmetic.
// Using a pointer alias allows nil checks and avoids copying large structs.
type Decimal = *apd.Decimal

// decimalContext is the default context for decimal operations with 34-digit precision.
var decimalContext = apd.BaseContext.WithPrecision(34)

// NewDecimal creates a new Decimal from a string representation.
// Returns an error if the string cannot be parsed.
//
// Example:
//
//	price, err := domain.NewDecimal("50000.12345678")
func NewDecimal(s string) (Decimal, error) {
	d, _, err := apd.NewFromString(s)
	if err != nil {
		return nil, fmt.Errorf("invalid decimal string %q: %w", s, err)
	}
	return d, nil
}

// MustDecimal creates a new Decimal from a string, panicking on error.
// Use only for compile-time constants where the value is known to be valid.
//
// Example:
//
//	minPrice := domain.MustDecimal("0.01")
func MustDecimal(s string) Decimal {
	d, err := NewDecimal(s)
	if err != nil {
		panic(err)
	}
	return d
}

// NewDecimalFromInt creates a new Decimal from an int64 value.
//
// Example:
//
//	quantity := domain.NewDecimalFromInt(100)
func NewDecimalFromInt(i int64) Decimal {
	return apd.New(i, 0)
}

// NewDecimalFromFloat64 creates a new Decimal from a float64 value.
// WARNING: Float64 has inherent precision limitations. Use string parsing
// when precision is critical (e.g., from exchange API responses).
//
// Example:
//
//	rate := domain.NewDecimalFromFloat64(0.001)
func NewDecimalFromFloat64(f float64) Decimal {
	d := apd.New(0, 0)
	d.SetFloat64(f)
	return d
}

// Zero returns a Decimal representing zero (0).
func Zero() Decimal {
	return apd.New(0, 0)
}

// One returns a Decimal representing one (1).
func One() Decimal {
	return apd.New(1, 0)
}

// Ten returns a Decimal representing ten (10).
func Ten() Decimal {
	return apd.New(10, 0)
}

// NegOne returns a Decimal representing negative one (-1).
func NegOne() Decimal {
	return apd.New(-1, 0)
}

// Add returns the sum of two Decimals (a + b).
// Returns a new Decimal, does not modify inputs.
func Add(a, b Decimal) Decimal {
	result := apd.New(0, 0)
	_, err := decimalContext.Add(result, a, b)
	if err != nil {
		panic(fmt.Sprintf("decimal add error: %v", err))
	}
	return result
}

// Sub returns the difference of two Decimals (a - b).
// Returns a new Decimal, does not modify inputs.
func Sub(a, b Decimal) Decimal {
	result := apd.New(0, 0)
	_, err := decimalContext.Sub(result, a, b)
	if err != nil {
		panic(fmt.Sprintf("decimal sub error: %v", err))
	}
	return result
}

// Mul returns the product of two Decimals (a * b).
// Returns a new Decimal, does not modify inputs.
func Mul(a, b Decimal) Decimal {
	result := apd.New(0, 0)
	_, err := decimalContext.Mul(result, a, b)
	if err != nil {
		panic(fmt.Sprintf("decimal mul error: %v", err))
	}
	return result
}

// Div returns the quotient of two Decimals (a / b).
// Returns a new Decimal, does not modify inputs.
// Panics if b is zero.
func Div(a, b Decimal) Decimal {
	if IsZero(b) {
		panic("decimal division by zero")
	}
	result := apd.New(0, 0)
	_, err := decimalContext.Quo(result, a, b)
	if err != nil {
		panic(fmt.Sprintf("decimal div error: %v", err))
	}
	return result
}

// QuoRem returns the quotient and remainder of a/b.
// The quotient is computed first, then remainder = a - (quotient * b).
// Returns (quotient, remainder, error).
func QuoRem(a, b Decimal, scale uint32) (Decimal, Decimal, error) {
	if IsZero(b) {
		return nil, nil, fmt.Errorf("division by zero")
	}
	// Compute quotient
	quo := Div(a, b)
	// Round quotient to specified scale
	quo = Round(quo, scale)
	// Compute remainder: a - (quo * b)
	rem := Sub(a, Mul(quo, b))
	return quo, rem, nil
}

// Mod returns the remainder of a/b (modulo operation).
func Mod(a, b Decimal) Decimal {
	if IsZero(b) {
		panic("decimal modulo by zero")
	}
	result := apd.New(0, 0)
	_, err := decimalContext.Rem(result, a, b)
	if err != nil {
		panic(fmt.Sprintf("decimal mod error: %v", err))
	}
	return result
}

// Abs returns the absolute value of a Decimal.
func Abs(d Decimal) Decimal {
	result := apd.New(0, 0)
	_, err := decimalContext.Abs(result, d)
	if err != nil {
		panic(fmt.Sprintf("decimal abs error: %v", err))
	}
	return result
}

// Neg returns the negation of a Decimal (-d).
func Neg(d Decimal) Decimal {
	result := apd.New(0, 0)
	_, err := decimalContext.Neg(result, d)
	if err != nil {
		panic(fmt.Sprintf("decimal neg error: %v", err))
	}
	return result
}

// Cmp compares two Decimals and returns:
//
//	-1 if a < b
//	 0 if a == b
//	+1 if a > b
func Cmp(a, b Decimal) int {
	return a.Cmp(b)
}

// Compare is an alias for Cmp for consistency with standard library.
func Compare(a, b Decimal) int {
	return Cmp(a, b)
}

// Equal returns true if two Decimals are equal.
func Equal(a, b Decimal) bool {
	return Cmp(a, b) == 0
}

// IsZero returns true if the Decimal equals zero.
func IsZero(d Decimal) bool {
	return d.IsZero()
}

// IsPositive returns true if the Decimal is greater than zero.
func IsPositive(d Decimal) bool {
	return Cmp(d, Zero()) > 0
}

// IsNegative returns true if the Decimal is less than zero.
func IsNegative(d Decimal) bool {
	return Cmp(d, Zero()) < 0
}

// IsInteger returns true if the Decimal has no fractional part.
func IsInteger(d Decimal) bool {
	// A number is an integer if exponent >= 0
	// Or if the coefficient when reduced has no fractional part
	if d.Exponent >= 0 {
		return true
	}
	// Check if the number is whole by checking if Modf returns zero fractional part
	integ, frac := apd.New(0, 0), apd.New(0, 0)
	d.Modf(integ, frac)
	return frac.IsZero()
}

// String returns the string representation of a Decimal.
func String(d Decimal) string {
	return d.String()
}

// StringFixed returns a string with the specified number of decimal places.
func StringFixed(d Decimal, precision uint32) string {
	rounded := Round(d, precision)
	return rounded.String()
}

// Float64 returns the float64 representation of a Decimal.
// WARNING: This loses precision for very large or very small numbers.
// Use only for display or non-critical calculations.
func Float64(d Decimal) (float64, error) {
	f, err := d.Float64()
	if err != nil {
		return 0, fmt.Errorf("cannot convert decimal to float64: %w", err)
	}
	return f, nil
}

// Int64 returns the int64 representation of a Decimal.
// Returns an error if the Decimal cannot be exactly represented as int64.
func Int64(d Decimal) (int64, error) {
	i, err := d.Int64()
	if err != nil {
		return 0, fmt.Errorf("cannot convert decimal to int64: %w", err)
	}
	return i, nil
}

// Clone returns a deep copy of a Decimal.
// Use this when returning copies of internal state.
func Clone(d Decimal) Decimal {
	if d == nil {
		return nil
	}
	result := apd.New(0, 0)
	result.Set(d)
	return result
}

// Min returns the smaller of two Decimals.
func Min(a, b Decimal) Decimal {
	if Cmp(a, b) <= 0 {
		return Clone(a)
	}
	return Clone(b)
}

// Max returns the larger of two Decimals.
func Max(a, b Decimal) Decimal {
	if Cmp(a, b) >= 0 {
		return Clone(a)
	}
	return Clone(b)
}

// Round rounds a Decimal to the specified number of decimal places.
func Round(d Decimal, precision uint32) Decimal {
	result := apd.New(0, 0)
	_, err := decimalContext.Quantize(result, d, -int32(precision))
	if err != nil {
		panic(fmt.Sprintf("decimal round error: %v", err))
	}
	return result
}

// Trunc truncates a Decimal to the specified number of decimal places (rounds toward zero).
func Trunc(d Decimal, precision uint32) Decimal {
	result := apd.New(0, 0)
	// Use Floor for positive, Ceil for negative to achieve truncation toward zero
	if !d.Negative {
		decimalContext.Floor(result, d)
	} else {
		decimalContext.Ceil(result, d)
	}
	// Adjust to the specified precision
	if precision > 0 {
		ctx := decimalContext.WithPrecision(precision + 10) // Extra precision to avoid rounding
		ctx.Quantize(result, result, -int32(precision))
	}
	return result
}
