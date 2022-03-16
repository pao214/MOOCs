## problem 1c

AND and OR gates alone cannot send a binary number containing all 0s to a 1

## problem 1d

Assume a function that
- maps the number with all 1s, say m, to a 0
- maps another number with atleast one 0 and one 1, say n, to 1
- maps the bitwise not of the above number, say p, to a 0
- xor of n and p should map to xor of 1 and 0 since the gates are linear
- however xor of n and p is m which maps to a 0

## problem 2c

the number of digits in exp(B, C) scales exponentially with n

## problem 2d

```
exp(B, C) mod D
  = ((exp(B, C/2) mod D) (exp(B, C/2) mod D)) mod D
  = ((exp(B, C/2) mod D) ^ 2) mod D

```

squaring takes O(n^2) since
- a number mod D cannot have more than n bits
- and from previous result we know that the multiplication of two n bit numbers can be done in O(n^2)

```
T(n) = T(n-1) + O(n^2)
```

## problem 4a

since NAND is universal, we can achieve signals equivalent to
- FLIP (1 - p)
- FLIP (p * q)

Subject to these constraints, the demoniator of FLIP prob. can only be a power of 2

## problem 4b

```
2^n mod 3
```

is always either 1 when n is even and 2 when n is odd

Consider function,

```
f: {0, 1}^n -> {0, 1}
```

for even n. This implies that the cardinality of the domain is 1 mod 3.

Each number in the set of all n-bit strings is equally likely to appear when we have n FLIP 1/2 signals.

Map all 1s n-bit string to `?`. Pr[?] = 1/2^n which can be made less than e with large enough n. 

3 evenly divides the reamining set and 1/3rd of the set can map to 1. Applying DNF we get FLIP 1/3.

And of all bits indicates a `?` in the second bit.

## problem 4d

The fermat primality test takes cosntant number of iterations for failure rate to reach below a target error.

This is in contrast to any deterministic algorithm whose number of iterations increases with higher numbers.

