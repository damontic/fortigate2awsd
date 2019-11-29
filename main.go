package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/cloudwatchlogs"
	"github.com/aws/aws-sdk-go/service/secretsmanager"
	"golang.org/x/crypto/ssh"
)

type fortigateCategory struct {
	id          int
	description string
}

func main() {
	version := "0.0.2"

	versionFlag := flag.Bool("version", false, "Set if you want to see the version and exit.")
	dryRun := flag.Bool("dry-run", false, "Set if you want to output messages to console. Useful for testing.")
	logGroupPrefix := flag.String("group-prefix", "", "Specify the log group prefix where you want to send the logs")
	logStream := flag.String("stream", "", "Specify the log stream where you want to send the logs")
	ipPort := flag.String("ip-port", "", "Specify the Fortigate ip and port to log to ip:port")
	username := flag.String("username", "", "Specify the Fortigate ssh username")
	password := flag.String("password", "", "Specify the Fortigate ssh password")
	secret := flag.String("secret-manager", "", "Specify the AWS secrets manager secrets name to use as password")
	eventSize := flag.Int("size", 100, "Specify the number of events to send to AWS Cloudwatch.")
	period := flag.Int("period", 300, "Specify the number of seconds to wait between logs category pushes.")
	verbose := flag.Bool("verbose", false, "Set if you want to be verbose.")
	flag.Parse()

	if *versionFlag {
		fmt.Println(version)
		os.Exit(0)
	}

	if !*dryRun && (*logGroupPrefix == "" || *logStream == "") {
		log.Fatalf("You must specify both the log group prefix and the log stream.\nCurrent logGroupPrefix: %s\nCurrent logStream: %s\nSee %s -h for help.", *logGroupPrefix, *logStream, os.Args[0])
	}

	if *ipPort == "" || *username == "" {
		log.Fatalf("You must specify:\n\t-ip-port as a string with format 'ip:port' where Fortigate is running.\n\t-username to use to connect to the Fortigate instance.\n\nA normal execution could be:\n\t%s -ip-port 192.168.0.1:22 -username admin -password 'a_secret_pass' -group cloudwatch_log_group -stream cloudwatch_log_stream\n\t%s -ip-port 192.168.0.1:22 -username admin -secret-manager 'aws_secret_manager_name' -group cloudwatch_log_group -stream cloudwatch_log_stream\n\nSee %s -h for help.", os.Args[0], os.Args[0], os.Args[0])
	}

	if *secret == "" && *password == "" {
		log.Fatalf("You must specify one of:\n\t-password 'a_password'\t(NOT RECOMENDED)\n\t-secret-manager 'an_aws_secret_manager_entry'\n\nSee %s -h for help.", os.Args[0])
	}

	fortigate2awsd(dryRun, period, eventSize, logGroupPrefix, logStream, ipPort, username, password, secret, verbose)
}

func fortigate2awsd(dryRun *bool, period, eventSize *int, logGroupPrefix, logStream, ipPort, username, password, secret *string, verbose *bool) {

	mySession := session.Must(session.NewSession())
	if *secret != "" {
		password = getSecretFromAwsSecretManager(mySession, secret)
	}

	cloudwatchlogsClient := cloudwatchlogs.New(mySession)

	config := &ssh.ClientConfig{
		User: *username,
		Auth: []ssh.AuthMethod{
			ssh.Password(*password),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}
	sshClient, err := ssh.Dial("tcp", *ipPort, config)
	if err != nil {
		log.Fatalf("Error in getSshWriteCloserAndScanner during ssh.Dial\n%v\n", err)
	}

	categories := []fortigateCategory{
		fortigateCategory{0, "traffic"},
		fortigateCategory{1, "event"},
		fortigateCategory{2, "virus"},
		fortigateCategory{3, "webfilter"},
		fortigateCategory{4, "ips"},
		fortigateCategory{5, "emailfilter"},
		fortigateCategory{7, "anomaly"},
		fortigateCategory{8, "voip"},
		fortigateCategory{9, "dlp"},
		fortigateCategory{10, "app-ctrl"},
		fortigateCategory{12, "waf"},
		fortigateCategory{15, "dns"},
		fortigateCategory{16, "ssh"},
		fortigateCategory{17, "ssl"},
		fortigateCategory{18, "cifs"},
		fortigateCategory{19, "file-filter"},
	}
	for {
		for _, category := range categories {
			if *verbose {
				log.Printf("Sending category: %s\n", category.description)
			}
			session, wc, scanner := getSshSessionWriteCloserAndScanner(sshClient)
			getFortigateLogsByCategory(*eventSize, category, wc, scanner, dryRun, cloudwatchlogsClient, logGroupPrefix, logStream, verbose)
			wc.Close()
			session.Close()
		}
		time.Sleep(time.Duration(*period) * time.Second)
	}

}

func getSecretFromAwsSecretManager(mySession *session.Session, secret *string) *string {
	secretsmanagerClient := secretsmanager.New(mySession)
	getSecretValueInput := &secretsmanager.GetSecretValueInput{
		SecretId: secret,
	}
	getSecretValueResult, err := secretsmanagerClient.GetSecretValue(getSecretValueInput)
	if err != nil {
		log.Fatalf("Error in sshClient during secretsmanagerClient.GetSecretValue\n%v\n", err)
	}
	return getSecretValueResult.SecretString
}

func getSshSessionWriteCloserAndScanner(sshClient *ssh.Client) (*ssh.Session, io.WriteCloser, *bufio.Scanner) {
	session, err := sshClient.NewSession()
	if err != nil {
		log.Fatalf("Error in getSshWriteCloserAndScanner during client.NewSession\n%v\n", err)
	}

	wc, err := session.StdinPipe()
	if err != nil {
		log.Fatalf("Error in getSshWriteCloserAndScanner during session.StdinPipe\n%v\n", err)
	}

	r, err := session.StdoutPipe()
	if err != nil {
		log.Fatalf("Error in getSshWriteCloserAndScanner during session.StdoutPipe\n%v\n", err)
	}

	scanner := bufio.NewScanner(r)
	scanner.Split(bufio.ScanLines)

	err = session.Shell()
	if err != nil {
		log.Fatalf("Error in getSshWriteCloserAndScanner during session.Shell\n%v\n", err)
	}

	return session, wc, scanner
}

func getFortigateLogsByCategory(eventSize int, category fortigateCategory, wc io.WriteCloser, scanner *bufio.Scanner, dryRun *bool, cloudwatchlogsClient *cloudwatchlogs.CloudWatchLogs, logGroupPrefix, logStream *string, verbose *bool) {
	logGroupS := *logGroupPrefix + "-" + category.description
	logGroup := &logGroupS

	if _, err := wc.Write([]byte("execute log filter device 1\n")); err != nil {
		log.Fatalf("Failed to run: log filter device 1\n%s\n", err.Error())
	}

	cmd := fmt.Sprintf("execute log filter category %d\n", category.id)
	if _, err := wc.Write([]byte(cmd)); err != nil {
		log.Fatalf("Failed to run: log filter category %d\n%s\n", category.id, err.Error())
	}
	if _, err := wc.Write([]byte("execute log filter start-line 1\n")); err != nil {
		log.Fatalf("Failed to run: log filter start-line 1\n%s\n", err.Error())
	}
	if _, err := wc.Write([]byte(fmt.Sprintf("execute log filter view-lines %d\n", eventSize))); err != nil {
		log.Fatalf("Failed to run: log filter start-line 1\n%s\n", err.Error())
	}
	if _, err := wc.Write([]byte("execute log display\n")); err != nil {
		log.Fatalf("Failed to run: log display\n%s\n", err.Error())
	}

	var events = make([]*cloudwatchlogs.InputLogEvent, 0)
	var nextToken *string
	var err error

	for i := 0; i < eventSize; i++ {
		if _, err := wc.Write([]byte("a\n")); err != nil {
			log.Fatalf("Failed to run: New Line to unblock Scan\n%s\n", err.Error())
		}
		scanner.Scan()
		m := scanner.Text()
		if len(m) > 50 {

			if !*dryRun {
				if *verbose {
					log.Printf("%s", m)
				}
				message, timestamp := getMessageTimestamp(m)

				event := &cloudwatchlogs.InputLogEvent{
					Message:   &message,
					Timestamp: &timestamp,
				}
				events = append(events, event)
			} else {
				fmt.Println(m)
			}
		}
	}

	if len(events) > 0 {
		sort.Sort(byTimestamp(events))
		nextToken, err = sendEventsCloudwatch(events, logGroup, logStream, nextToken, cloudwatchlogsClient)
		if err != nil {
			firstErrorLine := strings.Split(err.Error(), "\n")[0]
			splittedError := strings.Split(firstErrorLine, " ")
			nextToken, err = sendEventsCloudwatch(events, logGroup, logStream, &splittedError[len(splittedError)-1], cloudwatchlogsClient)
			if err != nil {
				log.Fatalf("%v", err)
			}
		}
	}
}

func getMessageTimestamp(m string) (string, int64) {
	re := regexp.MustCompile(`^.* eventtime=(.*) .*$`)
	result := re.FindStringSubmatchIndex(m)

	secondsPart := m[result[2] : result[2]+10]
	millisecondsPart := "000"
	millisecondsString := secondsPart + millisecondsPart

	milliseconds, err := strconv.ParseInt(millisecondsString, 10, 64)
	if err != nil {
		log.Fatalf("Error in getMessageTimestamp\n%v\nMessage: %s\ntimestamp: %s\n", m, millisecondsString, err)
	}

	return m, milliseconds
}

func sendEventsCloudwatch(events []*cloudwatchlogs.InputLogEvent, logGroupName *string, logStreamName *string, nextToken *string, cloudwatchlogsClient *cloudwatchlogs.CloudWatchLogs) (*string, error) {
	putLogEventInput := &cloudwatchlogs.PutLogEventsInput{
		LogEvents:     events,
		LogGroupName:  logGroupName,
		LogStreamName: logStreamName,
		SequenceToken: nextToken,
	}
	putLogEventsOutput, err := cloudwatchlogsClient.PutLogEvents(putLogEventInput)
	return putLogEventsOutput.NextSequenceToken, err
}

type byTimestamp []*cloudwatchlogs.InputLogEvent

func (a byTimestamp) Len() int           { return len(a) }
func (a byTimestamp) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a byTimestamp) Less(i, j int) bool { return *a[i].Timestamp < *a[j].Timestamp }
