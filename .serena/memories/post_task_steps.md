# Post-Task Completion Steps

Before submitting a change, ensure the following steps are performed:
1. Run `just fmt` to format the code.
2. Run `just vet` to check for common errors.
3. Run `just test` to ensure no regressions.
4. If database queries were modified, run `just sqlc-generate`.
5. Run `just test-coverage-summary` to check impact on coverage.
