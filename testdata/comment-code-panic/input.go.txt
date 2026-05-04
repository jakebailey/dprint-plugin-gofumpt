package commentcodepanic

// The combined constraint of an intersection type is the intersection of the constraints of
// the constituents. When an intersection type contains instantiable types with union type
// constraints, there are situations where we need to examine the combined constraint. One is
// when the target is a union type. Another is when the intersection contains types belonging
// to one of the disjoint domains. For example, given type variables T and U, each with the
// constraint 'string | number', the combined constraint of 'T & U' is 'string | number' and
// we need to check this constraint against a union on the target side. Also, given a type
// variable V constrained to 'string | number', 'V & number' has a combined constraint of
// 'string & number | number & number' which reduces to just 'number'.
// This also handles type parameters, as a type parameter with a union constraint compared against a union
// needs to have its constraint hoisted into an intersection with said type parameter, this way
// the type param can be compared with itself in the target (with the influence of its constraint to match other parts)
// For example, if `T extends 1 | 2` and `U extends 2 | 3` and we compare `T & U` to `T & U & (1 | 2 | 3)`

func f() {}
