package mat64

import (
	"math"

	"github.com/gonum/blas"
	"github.com/gonum/blas/blas64"
	"github.com/gonum/lapack/lapack64"
	"github.com/gonum/matrix"
)

var (
	triDense *TriDense
	_        Matrix        = triDense
	_        Triangular    = triDense
	_        RawTriangular = triDense
)

const badTriCap = "mat64: bad capacity for TriDense"

// TriDense represents an upper or lower triangular matrix in dense storage
// format.
type TriDense struct {
	mat blas64.Triangular
	cap int
}

type Triangular interface {
	Matrix
	// Triangular returns the number of rows/columns in the matrix and its
	// orientation.
	Triangle() (n int, kind matrix.TriKind)

	// TTri is the equivalent of the T() method in the Matrix interface but
	// guarantees the transpose is of triangular type.
	TTri() Triangular
}

type RawTriangular interface {
	RawTriangular() blas64.Triangular
}

var (
	_ Matrix           = TransposeTri{}
	_ Triangular       = TransposeTri{}
	_ UntransposeTrier = TransposeTri{}
)

// TransposeTri is a type for performing an implicit transpose of a Triangular
// matrix. It implements the Triangular interface, returning values from the
// transpose of the matrix within.
type TransposeTri struct {
	Triangular Triangular
}

// At returns the value of the element at row i and column j of the transposed
// matrix, that is, row j and column i of the Triangular field.
func (t TransposeTri) At(i, j int) float64 {
	return t.Triangular.At(j, i)
}

// Dims returns the dimensions of the transposed matrix. Triangular matrices are
// square and thus this is the same size as the original Triangular.
func (t TransposeTri) Dims() (r, c int) {
	c, r = t.Triangular.Dims()
	return r, c
}

// T performs an implicit transpose by returning the Triangular field.
func (t TransposeTri) T() Matrix {
	return t.Triangular
}

// Triangle returns the number of rows/columns in the matrix and its orientation.
func (t TransposeTri) Triangle() (int, matrix.TriKind) {
	n, upper := t.Triangular.Triangle()
	return n, !upper
}

// TTri performs an implicit transpose by returning the Triangular field.
func (t TransposeTri) TTri() Triangular {
	return t.Triangular
}

// Untranspose returns the Triangular field.
func (t TransposeTri) Untranspose() Matrix {
	return t.Triangular
}

func (t TransposeTri) UntransposeTri() Triangular {
	return t.Triangular
}

// NewTriDense creates a new Triangular matrix with n rows and columns. If data == nil,
// a new slice is allocated for the backing slice. If len(data) == n*n, data is
// used as the backing slice, and changes to the elements of the returned TriDense
// will be reflected in data. If neither of these is true, NewTriDense will panic.
//
// The data must be arranged in row-major order, i.e. the (i*c + j)-th
// element in the data slice is the {i, j}-th element in the matrix.
// Only the values in the triangular portion corresponding to kind are used.
func NewTriDense(n int, kind matrix.TriKind, data []float64) *TriDense {
	if n < 0 {
		panic("mat64: negative dimension")
	}
	if data != nil && len(data) != n*n {
		panic(matrix.ErrShape)
	}
	if data == nil {
		data = make([]float64, n*n)
	}
	uplo := blas.Lower
	if kind == matrix.Upper {
		uplo = blas.Upper
	}
	return &TriDense{
		mat: blas64.Triangular{
			N:      n,
			Stride: n,
			Data:   data,
			Uplo:   uplo,
			Diag:   blas.NonUnit,
		},
		cap: n,
	}
}

func (t *TriDense) Dims() (r, c int) {
	return t.mat.N, t.mat.N
}

// Triangle returns the dimension of t and its orientation. The returned
// orientation is only valid when n is not zero.
func (t *TriDense) Triangle() (n int, kind matrix.TriKind) {
	return t.mat.N, matrix.TriKind(!t.isZero()) && t.triKind()
}

func (t *TriDense) isUpper() bool {
	return isUpperUplo(t.mat.Uplo)
}

func (t *TriDense) triKind() matrix.TriKind {
	return matrix.TriKind(isUpperUplo(t.mat.Uplo))
}

func isUpperUplo(u blas.Uplo) bool {
	switch u {
	case blas.Upper:
		return true
	case blas.Lower:
		return false
	default:
		panic(badTriangle)
	}
}

// asSymBlas returns the receiver restructured as a blas64.Symmetric with the
// same backing memory. Panics if the receiver is unit.
// This returns a blas64.Symmetric and not a *SymDense because SymDense can only
// be upper triangular.
func (t *TriDense) asSymBlas() blas64.Symmetric {
	if t.mat.Diag == blas.Unit {
		panic("mat64: cannot convert unit TriDense into blas64.Symmetric")
	}
	return blas64.Symmetric{
		N:      t.mat.N,
		Stride: t.mat.Stride,
		Data:   t.mat.Data,
		Uplo:   t.mat.Uplo,
	}
}

// T performs an implicit transpose by returning the receiver inside a Transpose.
func (t *TriDense) T() Matrix {
	return Transpose{t}
}

// TTri performs an implicit transpose by returning the receiver inside a TransposeTri.
func (t *TriDense) TTri() Triangular {
	return TransposeTri{t}
}

func (t *TriDense) RawTriangular() blas64.Triangular {
	return t.mat
}

// Reset zeros the dimensions of the matrix so that it can be reused as the
// receiver of a dimensionally restricted operation.
//
// See the Reseter interface for more information.
func (t *TriDense) Reset() {
	// N and Stride must be zeroed in unison.
	t.mat.N, t.mat.Stride = 0, 0
	// Defensively zero Uplo to ensure
	// it is set correctly later.
	t.mat.Uplo = 0
	t.mat.Data = t.mat.Data[:0]
}

func (t *TriDense) isZero() bool {
	// It must be the case that t.Dims() returns
	// zeros in this case. See comment in Reset().
	return t.mat.Stride == 0
}

// untranspose untransposes a matrix if applicable. If a is an Untransposer, then
// untranspose returns the underlying matrix and true. If it is not, then it returns
// the input matrix and false.
func untransposeTri(a Triangular) (Triangular, bool) {
	if ut, ok := a.(UntransposeTrier); ok {
		return ut.UntransposeTri(), true
	}
	return a, false
}

// reuseAs resizes a zero receiver to an n??n triangular matrix with the given
// orientation. If the receiver is non-zero, reuseAs checks that the receiver
// is the correct size and orientation.
func (t *TriDense) reuseAs(n int, kind matrix.TriKind) {
	ul := blas.Lower
	if kind == matrix.Upper {
		ul = blas.Upper
	}
	if t.mat.N > t.cap {
		panic(badTriCap)
	}
	if t.isZero() {
		t.mat = blas64.Triangular{
			N:      n,
			Stride: n,
			Diag:   blas.NonUnit,
			Data:   use(t.mat.Data, n*n),
			Uplo:   ul,
		}
		t.cap = n
		return
	}
	if t.mat.N != n {
		panic(matrix.ErrShape)
	}
	if t.mat.Uplo != ul {
		panic(matrix.ErrTriangle)
	}
}

// isolatedWorkspace returns a new TriDense matrix w with the size of a and
// returns a callback to defer which performs cleanup at the return of the call.
// This should be used when a method receiver is the same pointer as an input argument.
func (t *TriDense) isolatedWorkspace(a Triangular) (w *TriDense, restore func()) {
	n, kind := a.Triangle()
	if n == 0 {
		panic("zero size")
	}
	w = getWorkspaceTri(n, kind, false)
	return w, func() {
		t.Copy(w)
		putWorkspaceTri(w)
	}
}

// Copy makes a copy of elements of a into the receiver. It is similar to the
// built-in copy; it copies as much as the overlap between the two matrices and
// returns the number of rows and columns it copied. Only elements within the
// receiver's non-zero triangle are set.
//
// See the Copier interface for more information.
func (t *TriDense) Copy(a Matrix) (r, c int) {
	r, c = a.Dims()
	r = min(r, t.mat.N)
	c = min(c, t.mat.N)
	if r == 0 || c == 0 {
		return 0, 0
	}

	switch a := a.(type) {
	case RawMatrixer:
		amat := a.RawMatrix()
		if t.isUpper() {
			for i := 0; i < r; i++ {
				copy(t.mat.Data[i*t.mat.Stride+i:i*t.mat.Stride+c], amat.Data[i*amat.Stride+i:i*amat.Stride+c])
			}
		} else {
			for i := 0; i < r; i++ {
				copy(t.mat.Data[i*t.mat.Stride:i*t.mat.Stride+i+1], amat.Data[i*amat.Stride:i*amat.Stride+i+1])
			}
		}
	case RawTriangular:
		amat := a.RawTriangular()
		aIsUpper := isUpperUplo(amat.Uplo)
		tIsUpper := t.isUpper()
		switch {
		case tIsUpper && aIsUpper:
			for i := 0; i < r; i++ {
				copy(t.mat.Data[i*t.mat.Stride+i:i*t.mat.Stride+c], amat.Data[i*amat.Stride+i:i*amat.Stride+c])
			}
		case !tIsUpper && !aIsUpper:
			for i := 0; i < r; i++ {
				copy(t.mat.Data[i*t.mat.Stride:i*t.mat.Stride+i+1], amat.Data[i*amat.Stride:i*amat.Stride+i+1])
			}
		default:
			for i := 0; i < r; i++ {
				t.set(i, i, amat.Data[i*amat.Stride+i])
			}
		}
	default:
		isUpper := t.isUpper()
		for i := 0; i < r; i++ {
			if isUpper {
				for j := i; j < c; j++ {
					t.set(i, j, a.At(i, j))
				}
			} else {
				for j := 0; j <= i; j++ {
					t.set(i, j, a.At(i, j))
				}
			}
		}
	}

	return r, c
}

// InverseTri computes the inverse of the triangular matrix a, storing the result
// into the receiver. If a is ill-conditioned, a Condition error will be returned.
// Note that matrix inversion is numerically unstable, and should generally be
// avoided where possible, for example by using the Solve routines.
func (t *TriDense) InverseTri(a Triangular) error {
	if rt, ok := a.(RawTriangular); ok {
		t.checkOverlap(rt.RawTriangular())
	}
	n, _ := a.Triangle()
	t.reuseAs(a.Triangle())
	t.Copy(a)
	work := make([]float64, 3*n)
	iwork := make([]int, n)
	cond := lapack64.Trcon(matrix.CondNorm, t.mat, work, iwork)
	if math.IsInf(cond, 1) {
		return matrix.Condition(cond)
	}
	ok := lapack64.Trtri(t.mat)
	if !ok {
		return matrix.Condition(math.Inf(1))
	}
	if cond > matrix.ConditionTolerance {
		return matrix.Condition(cond)
	}
	return nil
}

// MulTri takes the product of triangular matrices a and b and places the result
// in the receiver. The size of a and b must match, and they both must have the
// same TriKind, or Mul will panic.
func (t *TriDense) MulTri(a, b Triangular) {
	n, kind := a.Triangle()
	nb, kindb := b.Triangle()
	if n != nb {
		panic(matrix.ErrShape)
	}
	if kind != kindb {
		panic(matrix.ErrTriangle)
	}

	aU, _ := untransposeTri(a)
	bU, _ := untransposeTri(b)
	t.reuseAs(n, kind)
	var restore func()
	if t == aU {
		t, restore = t.isolatedWorkspace(aU)
		defer restore()
	} else if t == bU {
		t, restore = t.isolatedWorkspace(bU)
		defer restore()
	}

	// TODO(btracey): Improve the set of fast-paths.
	if kind == matrix.Upper {
		for i := 0; i < n; i++ {
			for j := i; j < n; j++ {
				var v float64
				for k := i; k <= j; k++ {
					v += a.At(i, k) * b.At(k, j)
				}
				t.SetTri(i, j, v)
			}
		}
		return
	}
	for i := 0; i < n; i++ {
		for j := 0; j <= i; j++ {
			var v float64
			for k := j; k <= i; k++ {
				v += a.At(i, k) * b.At(k, j)
			}
			t.SetTri(i, j, v)
		}
	}
}

// copySymIntoTriangle copies a symmetric matrix into a TriDense
func copySymIntoTriangle(t *TriDense, s Symmetric) {
	n, upper := t.Triangle()
	ns := s.Symmetric()
	if n != ns {
		panic("mat64: triangle size mismatch")
	}
	ts := t.mat.Stride
	if rs, ok := s.(RawSymmetricer); ok {
		sd := rs.RawSymmetric()
		ss := sd.Stride
		if upper {
			if sd.Uplo == blas.Upper {
				for i := 0; i < n; i++ {
					copy(t.mat.Data[i*ts+i:i*ts+n], sd.Data[i*ss+i:i*ss+n])
				}
				return
			}
			for i := 0; i < n; i++ {
				for j := i; j < n; j++ {
					t.mat.Data[i*ts+j] = sd.Data[j*ss+i]
				}
				return
			}
		}
		if sd.Uplo == blas.Upper {
			for i := 0; i < n; i++ {
				for j := 0; j <= i; j++ {
					t.mat.Data[i*ts+j] = sd.Data[j*ss+i]
				}
			}
			return
		}
		for i := 0; i < n; i++ {
			copy(t.mat.Data[i*ts:i*ts+i+1], sd.Data[i*ss:i*ss+i+1])
		}
		return
	}
	if upper {
		for i := 0; i < n; i++ {
			for j := i; j < n; j++ {
				t.mat.Data[i*ts+j] = s.At(i, j)
			}
		}
		return
	}
	for i := 0; i < n; i++ {
		for j := 0; j <= i; j++ {
			t.mat.Data[i*ts+j] = s.At(i, j)
		}
	}
}
