package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"regexp"
	"time"

	v1 "k8s.io/api/core/v1"
)

type Arguments struct {
	kubeconfig  string
	sessionFile string
	api_id      int
	api_hash    string
	persist     bool
	removeAll   bool
	resultsFile string
	messageIDs  string
	chatID      int
	imageName   string
	session     string
	namespace   string
}

// Read content of pyrogram session file
func readSessionContent(sessionPath string) string {
	file, err := os.Open(sessionPath)
	if err != nil {
		fmt.Printf("Error: opening file: %v\n", err)
		os.Exit(1)
	}
	defer file.Close() // Make sure to close the file when done

	content, err := io.ReadAll(file)
	if err != nil {
		fmt.Printf("Error: reading file: %v\n", err)
		os.Exit(1)
	}

	return string(content)
}

func getArguments() Arguments {
	// Let's read all variables from the command line args
	var arguments Arguments

	flag.StringVar(&arguments.kubeconfig, "kubeconfig", "", "(!) Path to the kubeconfig file")
	flag.StringVar(&arguments.sessionFile, "session", "", "(!) Path to the file with a telegram Pyrogram session string")
	flag.IntVar(&arguments.api_id, "api-id", 0, "(!) integer Telegram API id")
	flag.StringVar(&arguments.api_hash, "api-hash", "", "(!) Telegram API hash")
	flag.BoolVar(&arguments.persist, "persist", false, "(Opt.) Create a persistent volume and export results to it")
	flag.BoolVar(&arguments.removeAll, "remove-all", false, "(Opt.)Remove messages all messages in public chats")
	flag.StringVar(&arguments.resultsFile, "results-filename", "results_"+time.Now().Format("20060102150405")+".csv", "(Opt.) Specify the results filename")
	flag.StringVar(&arguments.messageIDs, "message-ids", "", "(Opt.) Specify message IDs as a integers array with no spaces e.g. [1,2,3,4]")
	flag.IntVar(&arguments.chatID, "chat-id", 0, "(Opt.) Specify chat ID to remove messages in it")
	flag.StringVar(&arguments.imageName, "imagename", "you2ku/python-tg-script:latest", "(Opt.) Custom image name. Defaults to you2ku/python-tg-script:latest")
	flag.StringVar(&arguments.namespace, "namespace", "default", "(Opt.) k8s namespace")

	flag.Parse()

	// Providing information to the end-user

	fmt.Println("deploy: v1. For help please use --help")

	if flag.Arg(0) == "--help" {
		fmt.Println("Deploy Python Telegram Message Management script docker image to Kubernetes and attach to logs:")
		flag.Usage()
		os.Exit(0)
	}

	// Validating inputs
	if arguments.api_id == 0 {
		fmt.Println("Error: -api_id is required!")
		os.Exit(1)
	}

	if arguments.api_hash == "" {
		fmt.Println("Error: -api_hash is required!")
		os.Exit(1)
	}

	if arguments.sessionFile == "" {
		fmt.Println("Error: -session path to file is required!")
		os.Exit(1)
	} else {
		arguments.session = readSessionContent(arguments.sessionFile)
	}

	if arguments.kubeconfig == "" {
		fmt.Println("Error: -kubeconfig path to file is required!")
		os.Exit(1)
	}

	check_messageIDs := regexp.MustCompile(`^\d*(,\d+)*$`)

	if !check_messageIDs.MatchString(arguments.messageIDs) {
		fmt.Println("Error: -message-ids have to be a list of ints, comma separated, no spaces.")
		os.Exit(1)
	}

	fmt.Println("Values of command-line flags:")
	fmt.Println("sessionFile:", arguments.sessionFile)
	fmt.Println("persist:", arguments.persist)
	fmt.Println("removeAll:", arguments.removeAll)
	fmt.Println("resultsFile:", arguments.resultsFile)
	fmt.Println("messageIDs:", arguments.messageIDs)
	fmt.Println("chatID:", arguments.chatID)
	fmt.Println("kubeconfigPath:", arguments.kubeconfig)

	return arguments
}

func main() {
	arguments := getArguments()

	fmt.Println("Deploying docker image to Kubernetes...")
	fmt.Println("Reading configuration...")
	clientset := getClient(arguments)
	job_id := getJobId()

	fmt.Println("Running a job with job id:", job_id)
	var pod *v1.Pod
	if arguments.persist {
		vol := createPersistentVolume(clientset, job_id)
		fmt.Println("Created Persistent Volume:", vol.Name)
		fmt.Println("Running the main script:")
		pod = runMainScript(clientset, job_id, arguments, vol)
	} else {
		pod = runMainScript(clientset, job_id, arguments, nil)
	}

	watchPodStatus(clientset, arguments.namespace, pod.Name)

}
