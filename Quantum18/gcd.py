def gcd(a, b):
    if a < b:
        return gcd(b, a)

    if b == 0:
        return a

    return gcd(b, a%b)

print(gcd(106113609170668254652391269192197757215334846951209743863306173107325600, 894128743023837450195367301428030915619905056250057436341104142570200))
