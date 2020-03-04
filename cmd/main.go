package main

import (
	"log"
	"os"
	"os/exec"
)

func main() {

	//TODO this won't be needed in future, because it can possibly be fetched in cloudbuild?
	clone := exec.Command("git", "clone", "git@github.com:hromadkavojta/BP-infratest.git")
	err := clone.Run()
	if err != nil {
		log.Printf("Couldn't fetch github repo")
	}

	//Running terraform init
	tfInit := exec.Command("terraform", "init", "BP-infratest")
	tfInit.Stdout = os.Stdout
	tfInit.Stderr = os.Stderr
	tfInit.Run()
	if err != nil {
		log.Printf("Terraform couldnt run init")
	}

	//Running terraform plan (all these steps (excludet dir name, which should be ideally environment variable)
	cmd := exec.Command("terraform", "plan", "-out=plan.out", "BP-infratest")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	log.Printf("Running terminal command ECHO\n")
	err = cmd.Run()
	if err != nil {
		log.Printf("command Echo failed\n")
	}

	//Removing cloned repository ( this is not ideal because of huge data transmition --> primary meant cloning)
	rm_dir := exec.Command("rm", "-r", "BP-infratest")
	rm_dir.Run()
	if err != nil {
		log.Printf("Couldnt remove directory %v", err)
	}
}
