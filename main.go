package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"os"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/cloudwatchlogs"
	"golang.org/x/crypto/ssh"
)

func main() {
	version := "0.0.0"

	versionFlag := flag.Bool("version", false, "Set if you want to see the version and exit.")
	dryRun := flag.Bool("dry-run", false, "Set if you want to output messages to console. Useful for testing.")
	logGroup := flag.String("group", "", "Specify the log group where you want to send the logs")
	logStream := flag.String("stream", "", "Specify the log stream where you want to send the logs")
	ipPort := flag.String("ip-port", "", "Specify the Fortigate ip and port to log to ip:port")
	username := flag.String("username", "", "Specify the Fortigate ssh username")
	password := flag.String("password", "", "Specify the Fortigate ssh password")
	eventSize := flag.Int("size", 10, "Specify the number of events to send to AWS Cloudwatch.")
	flag.Parse()

	if *versionFlag {
		fmt.Println(version)
		os.Exit(0)
	}

	if !*dryRun && (*logGroup == "" || *logStream == "") {
		log.Fatalf("You must specify both the log group and the log stream.\nCurrent logGroup: %s\nCurrent logStream: %s\nSee %s -h for help.", *logGroup, *logStream, os.Args[0])
	}

	if *ipPort == "" || *username == "" || *password == "" {
		log.Fatalf("You must specify:\n\t-ip-port as a string with format 'ip:port' where Fortigate is running.\n\t-username to use to connect to the Fortigate instance.\n\t-password to use to connect to the Fortigate instance.\n\nA normal execution could be:\n\t%s -ip-port 192.168.0.1:22 -username admin -password 'a_secret_pass' -group cloudwatch_log_group -stream cloudwatch_log_stream\n\nSee %s -h for help.", os.Args[0], os.Args[0])
	}

	fortigate2awsd(dryRun, eventSize, logGroup, logStream, ipPort, username, password)
}

func fortigate2awsd(dryRun *bool, eventSize *int, logGroup, logStream, ipPort, username, password *string) {
	mySession := session.Must(session.NewSession())
	cloudwatchlogsClient := cloudwatchlogs.New(mySession)

	config := &ssh.ClientConfig{
		User: *username,
		Auth: []ssh.AuthMethod{
			ssh.Password(*password),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}
	client, err := ssh.Dial("tcp", *ipPort, config)
	if err != nil {
		log.Fatalf("Error in sshClient during ssh.Dial\n%v\n", err)
	}

	session, err := client.NewSession()
	if err != nil {
		log.Fatalf("Error in sshClient during client.NewSession\n%v\n", err)
	}
	defer session.Close()

	wc, err := session.StdinPipe()
	if err != nil {
		log.Fatalf("Error in sshClient during session.StdinPipe\n%v\n", err)
	}
	defer wc.Close()

	r, err := session.StdoutPipe()
	if err != nil {
		log.Fatalf("Error in sshClient during session.StdoutPipe\n%v\n", err)
	}

	scanner := bufio.NewScanner(r)
	scanner.Split(bufio.ScanLines)

	err = session.Shell()
	if err != nil {
		log.Fatalf("Error in sshClient during session.Shell\n%v\n", err)
	}

	if _, err := wc.Write([]byte("execute log filter device 1\n")); err != nil {
		log.Fatalf("Failed to run: log filter device 1\n%s\n", err.Error())
	}
	if _, err := wc.Write([]byte("execute log filter start-line 1\n")); err != nil {
		log.Fatalf("Failed to run: log filter start-line 1\n%s\n", err.Error())
	}
	if _, err := wc.Write([]byte("execute log filter category 0\n")); err != nil {
		log.Fatalf("Failed to run: log filter category 0\n%s\n", err.Error())
	}
	if _, err := wc.Write([]byte("execute log display\n")); err != nil {
		log.Fatalf("Failed to run: log display\n%s\n", err.Error())
	}

	var counter = 0
	var events = make([]*cloudwatchlogs.InputLogEvent, *eventSize)
	var nextToken *string

	for scanner.Scan() {
		m := scanner.Text()
		if len(m) > 50 {

			if !*dryRun {
				message, timestamp := getMessageTimestamp(m)

				events[counter] = &cloudwatchlogs.InputLogEvent{
					Message:   &message,
					Timestamp: &timestamp,
				}

				if counter == *eventSize-1 {
					counter = 0
					nextToken, err = sendEventsCloudwatch(events, logGroup, logStream, nextToken, cloudwatchlogsClient)
					if err != nil {
						firstErrorLine := strings.Split(err.Error(), "\n")[0]
						splittedError := strings.Split(firstErrorLine, " ")
						nextToken, err = sendEventsCloudwatch(events, logGroup, logStream, &splittedError[len(splittedError)-1], cloudwatchlogsClient)
						if err != nil {
							log.Fatalf("%v", err)
						}
					}
				} else {
					fmt.Println(m)
				}
				counter++
			}

			if _, err := wc.Write([]byte("execute log display\n")); err != nil {
				log.Fatalf("Failed to run: log display\n%s\n", err.Error())
			}
		}
	}
}

func sendEventsCloudwatch(events []*cloudwatchlogs.InputLogEvent, logGroupName *string, logStreamName *string, nextToken *string, cloudwatchlogsClient *cloudwatchlogs.CloudWatchLogs) (*string, error) {

	fmt.Printf("%v\n", events)

	sort.Sort(ByTimestamp(events))

	putLogEventInput := &cloudwatchlogs.PutLogEventsInput{
		LogEvents:     events,
		LogGroupName:  logGroupName,
		LogStreamName: logStreamName,
		SequenceToken: nextToken,
	}
	putLogEventsOutput, err := cloudwatchlogsClient.PutLogEvents(putLogEventInput)
	return putLogEventsOutput.NextSequenceToken, err
}

func sendEventsConsole(events []*cloudwatchlogs.InputLogEvent) {
	fmt.Printf("%v\n", events)
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

type ByTimestamp []*cloudwatchlogs.InputLogEvent

func (a ByTimestamp) Len() int           { return len(a) }
func (a ByTimestamp) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a ByTimestamp) Less(i, j int) bool { return *a[i].Timestamp < *a[j].Timestamp }
