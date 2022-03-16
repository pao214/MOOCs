## prob 4d

P = |psi><psi|
|x> = a|psi>+b|opsi>
where opsi is orthonormal to psi
P|x> = a|psi>
=> projection of |x> along |psi>

## prob 4e

1 - 2P => direction of projection reversed => reflection of qubit in the hyperplane perp to |psi>

## prob 6

<psi|psi> = <psi| At * A |psi> = <psi| B |psi> where B= At * A and At is A conjugate transpose of A
all of B's eigen values are 1 implies B is identity

## prob 8e

D > 1/2C => C-D < 1/2C and C mod D is atmost C-D
D <= 1/2C => CmodD < 1/2C since C mod D is atmost D-1

By case analysis Ft+1 <= 1/2(Ft-1)
=> Ft+1 Ft <= 1/2 Ft Ft-1

The value of Ft Ft+1 halves each time and hence the total number of steps to reach a constant is
  atmost log2(A B) where A and B are initial number

=> Log2 A + Log2 B
