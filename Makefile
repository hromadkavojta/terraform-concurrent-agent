create:
	go build -ldflags "-X main.Url=https://agent-dot-vojtah-sandbox.ew.r.appspot.com/ -X main.Hash=THISISRANDOMGENERATEDHASHINBUILD" -o tfagent client/main.go
	gcloud app deploy agent/app.yaml
