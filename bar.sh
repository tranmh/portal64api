/root/go/bin/swag init -g cmd/server/main.go -o docs/generated
make build
export MVDSB_USERNAME=portal
export MVDSB_PASSWORD='Usm@1?/#Qv^avF'
export MVDSB_HOST=localhost
export MVDSB_DATABASE=mvdsb
export PORTAL64_BDW_USERNAME=portal
export PORTAL64_BDW_PASSWORD='Usm@1?/#Qv^avF'
export PORTAL64_BDW_HOST=localhost
export PORTAL64_BDW_DATABASE=portal64_bdw
./bin/portal64api