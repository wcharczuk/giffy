NO_COLOR=\033[0m
OK_COLOR=\033[32;01m
ERROR_COLOR=\033[31;01m
WARN_COLOR=\033[33;01m

cover:
	@echo "$(OK_COLOR)Spiffy ==> Running Coverage$(NO_COLOR)"
	@go test -short -covermode=set -coverprofile=cover_profile.tmp
	@go tool cover -html=cover_profile.tmp
	@rm cover_profile.tmp
	@echo "$(OK_COLOR)Spiffy ==> Running Coverage Done!$(NO_COLOR)"

bench:
	@echo "$(OK_COLOR)Spiffy ==> Running Benchmarks$(NO_COLOR)"
	@go test -bench=.
	@echo "$(OK_COLOR)Spiffy ==> Running Benchmarks Done!$(NO_COLOR)"
