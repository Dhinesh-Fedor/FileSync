package cli

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"time"

	"FileSync/internal/client"
	"FileSync/internal/models"
	"FileSync/internal/scanFiles"
	serverapp "FileSync/internal/server"
)

var (
	Version   = "2.0.0"
	BuildDate = "unknown"
)

func Execute(args []string) error {
	args = normalizeArgs(args)
	if len(args) == 0 {
		printGeneralHelp()
		return nil
	}

	if args[0] == "help" {
		printHelp(args[1:])
		return nil
	}

	if isHelpFlag(args[0]) {
		printGeneralHelp()
		return nil
	}

	switch args[0] {
	case "version":
		printVersion()
		return nil
	case "add":
		return runAdd(args[1:])
	case "list":
		return runList()
	case "details":
		return runDetails(args[1:])
	case "delete":
		return runDelete(args[1:])
	case "scan":
		return runScan(args[1:])
	case "compare":
		return runCompare(args[1:])
	case "changes":
		return runChanges(args[1:])
	case "sync":
		return runSync(args[1:])
	case "status":
		return runStatus(args[1:])
	case "report":
		return runReport(args[1:])
	case "server":
		return runServer(args[1:])
	default:
		fmt.Println("Unknown command.")
		printGeneralHelp()
		return nil
	}
}

func normalizeArgs(args []string) []string {
	if len(args) > 0 && args[0] == "fs" {
		return args[1:]
	}
	return args
}

func isHelpFlag(arg string) bool {
	return arg == "-h" || arg == "--help" || arg == "help"
}

func printHelp(args []string) {
	if len(args) == 0 {
		printGeneralHelp()
		return
	}

	switch args[0] {
	case "add":
		printAddHelp()
	case "list":
		printListHelp()
	case "details":
		printDetailsHelp()
	case "delete":
		printDeleteHelp()
	case "scan":
		printScanHelp()
	case "compare":
		printCompareHelp()
	case "changes":
		printChangesHelp()
	case "sync":
		printSyncHelp()
	case "status":
		printStatusHelp()
	case "report":
		printReportHelp()
	case "server":
		printServerHelp()
	case "version":
		printVersionHelp()
	default:
		printGeneralHelp()
	}
}

func printGeneralHelp() {
	fmt.Println("------------------------------------------------")
	fmt.Println()
	fmt.Println("FileSync")
	fmt.Println()
	fmt.Println("A centralized file synchronization tool.")
	fmt.Println()
	fmt.Println("Usage")
	fmt.Println()
	fmt.Println("    filesync <command> [arguments]")
	fmt.Println()
	fmt.Println("Commands")
	fmt.Println()
	fmt.Println("Client")
	fmt.Println()
	fmt.Println("    add         Register a folder")
	fmt.Println("    list        List registered folders")
	fmt.Println("    details     Show folder details")
	fmt.Println("    delete      Remove a folder")
	fmt.Println("    scan        Scan local folder")
	fmt.Println("    compare     Compare metadata")
	fmt.Println("    changes     Show pending changes")
	fmt.Println("    sync        Synchronize with server")
	fmt.Println("    status      Show synchronization status")
	fmt.Println("    report      Show latest sync report")
	fmt.Println()
	fmt.Println("Server")
	fmt.Println()
	fmt.Println("    server start        Start FileSync server")
	fmt.Println("    server stop         Stop server")
	fmt.Println("    server status       Show server status")
	fmt.Println("    server logs         Show server logs")
	fmt.Println()
	fmt.Println("General")
	fmt.Println("    version     Show version")
	fmt.Println("    help        Show help")
	fmt.Println()
	fmt.Println("Examples")
	fmt.Println()
	fmt.Println("    filesync add")
	fmt.Println()
	fmt.Println("    filesync list")
	fmt.Println()
	fmt.Println("    filesync sync")
	fmt.Println()
	fmt.Println("    filesync server start")
	fmt.Println()
	fmt.Println("    filesync server status")
	fmt.Println()
	fmt.Println("    filesync report")
	fmt.Println()
	fmt.Println("Use")
	fmt.Println()
	fmt.Println("    filesync help <command>")
	fmt.Println()
	fmt.Println("for more information.")
	fmt.Println()
	fmt.Println("------------------------------------------------")
}

func printAddHelp() {
	printCommandHelp("Register a folder for synchronization.", "filesync add", []string{"Folder name", "Folder path", "Server URL"}, []string{"filesync add", "filesync add Docs /home/me/Documents http://localhost:8080"})
}

func printSyncHelp() {
	printCommandHelp("Synchronize a registered folder with the FileSync server.", "filesync sync <folder-id>", []string{"Folder ID or use interactive selection"}, []string{"filesync sync folder-1785874016247429487"})
}

func printServerHelp() {
	fmt.Println("Description")
	fmt.Println()
	fmt.Println("Manage the FileSync server.")
	fmt.Println()
	fmt.Println("Usage")
	fmt.Println()
	fmt.Println("filesync server <command>")
	fmt.Println()
	fmt.Println("Commands")
	fmt.Println()
	fmt.Println("start")
	fmt.Println("stop")
	fmt.Println("status")
	fmt.Println("logs")
	fmt.Println()
	fmt.Println("Examples")
	fmt.Println()
	fmt.Println("filesync server start")
	fmt.Println("filesync server status")
}

func printListHelp() {
	printCommandHelp("List registered folders.", "filesync list", []string{"None"}, []string{"filesync list"})
}

func printDetailsHelp() {
	printCommandHelp("Show folder details.", "filesync details <folder-id>", []string{"Folder ID or interactive selection"}, []string{"filesync details folder-1785874016247429487", "filesync details"})
}

func printDeleteHelp() {
	printCommandHelp("Remove a folder.", "filesync delete <folder-id>", []string{"Folder ID or interactive selection"}, []string{"filesync delete folder-1785874016247429487", "filesync delete"})
}

func printScanHelp() {
	printCommandHelp("Scan a local folder.", "filesync scan <folder-path>", []string{"Folder path or registered folder"}, []string{"filesync scan /tmp/src", "filesync scan Projects"})
}

func printCompareHelp() {
	printCommandHelp("Compare metadata for two folders.", "filesync compare <source> <destination>", []string{"Source path", "Destination path"}, []string{"filesync compare /tmp/src /tmp/dst"})
}

func printChangesHelp() {
	printCommandHelp("Show pending changes for a registered folder.", "filesync changes <folder-id>", []string{"Folder ID or interactive selection"}, []string{"filesync changes folder-1785874016247429487", "filesync changes"})
}

func printStatusHelp() {
	printCommandHelp("Show synchronization status.", "filesync status", []string{"None"}, []string{"filesync status"})
}

func printReportHelp() {
	printCommandHelp("Show the latest sync report.", "filesync report [folder-id]", []string{"Optional folder ID"}, []string{"filesync report", "filesync report folder-1785874016247429487"})
}

func printVersionHelp() {
	printCommandHelp("Show the FileSync version.", "filesync version", []string{"None"}, []string{"filesync version"})
}

func printCommandHelp(description string, usage string, arguments []string, examples []string) {
	fmt.Println("Description")
	fmt.Println()
	fmt.Println(description)
	fmt.Println()
	fmt.Println("Usage")
	fmt.Println()
	fmt.Println(usage)
	fmt.Println()
	fmt.Println("Arguments")
	fmt.Println()
	for _, arg := range arguments {
		fmt.Println("    " + arg)
	}
	fmt.Println()
	fmt.Println("Example")
	fmt.Println()
	for _, example := range examples {
		fmt.Println("    " + example)
	}
}

func printVersion() {
	fmt.Println("FileSync")
	fmt.Println("Version   :", Version)
	fmt.Println("Build Date:", BuildDate)
	fmt.Println("Go Version:", runtime.Version())
}

func runAdd(args []string) error {
	name, err := promptOrArg(args, 0, "Folder name")
	if err != nil {
		return err
	}
	pathValue, err := promptOrArg(args, 1, "Folder path")
	if err != nil {
		return err
	}
	serverURL, err := promptOrArg(args, 2, "Server URL")
	if err != nil {
		return err
	}

	if strings.TrimSpace(name) == "" || strings.TrimSpace(pathValue) == "" || strings.TrimSpace(serverURL) == "" {
		fmt.Println("Error: missing folder information.")
		printAddHelp()
		return nil
	}

	if _, err := os.Stat(pathValue); err != nil {
		if os.IsNotExist(err) {
			fmt.Println("Folder path does not exist.")
			return nil
		}
		return err
	}

	folder, err := client.RegisterFolder(name, pathValue, serverURL)
	if err != nil {
		return err
	}

	fmt.Printf("Folder registered: %s\n", folder.ID)
	return nil
}

func runList() error {
	folders, err := client.ListFolders()
	if err != nil {
		return err
	}

	if len(folders) == 0 {
		fmt.Println("No registered folders found.")
		return nil
	}

	fmt.Println("ID | NAME | PATH | VERSION")
	fmt.Println("----------------------------------------------")
	for _, folder := range folders {
		fmt.Printf("%s | %s | %s | %d\n", folder.ID, folder.Name, folder.LocalPath, folder.CurrentVersion)
	}

	return nil
}

func runDetails(args []string) error {
	folderID, err := resolveFolderID(args, "Folder ID")
	if err != nil {
		return err
	}
	if strings.TrimSpace(folderID) == "" {
		fmt.Println("Error: missing folder ID.")
		printDetailsHelp()
		return nil
	}

	folder, err := client.GetFolder(folderID)
	if err != nil {
		return err
	}
	if folder == nil {
		fmt.Println("Folder not found.")
		return nil
	}

	fmt.Println("------FOLDER DETAILS------")
	fmt.Printf("ID           : %s\n", folder.ID)
	fmt.Printf("Name         : %s\n", folder.Name)
	fmt.Printf("Local Path   : %s\n", folder.LocalPath)
	fmt.Printf("Remote ID    : %s\n", folder.RemoteID)
	fmt.Printf("Version      : %d\n", folder.CurrentVersion)
	fmt.Printf("Device ID    : %s\n", folder.DeviceID)
	fmt.Printf("Last Synced  : %s\n", folder.LastSyncedAt.Format(time.RFC3339))
	return nil
}

func runDelete(args []string) error {
	folderID, err := resolveFolderID(args, "Folder ID")
	if err != nil {
		return err
	}
	if strings.TrimSpace(folderID) == "" {
		fmt.Println("Error: missing folder ID.")
		printDeleteHelp()
		return nil
	}

	if err := client.DeleteFolder(folderID); err != nil {
		if strings.Contains(err.Error(), "folder not found") {
			fmt.Println("Folder not found.")
			return nil
		}
		return err
	}

	fmt.Printf("Folder %s deleted.\n", folderID)
	return nil
}

func runScan(args []string) error {
	if len(args) == 0 {
		folderPath, err := promptValue("Folder path")
		if err != nil {
			return err
		}
		if folderPath == "" {
			fmt.Println("Error: missing folder path.")
			printScanHelp()
			return nil
		}

		return runSingleScan(folderPath)
	}

	if len(args) == 1 {
		return runSingleScan(args[0])
	}

	return runTwoScan(args[0], args[1])
}

func runCompare(args []string) error {
	if len(args) < 2 {
		fmt.Println("Error: missing source or destination path.")
		printCompareHelp()
		return nil
	}

	return runComparison(args[0], args[1])
}

func runChanges(args []string) error {
	folderID, err := resolveFolderID(args, "Folder ID")
	if err != nil {
		return err
	}
	if strings.TrimSpace(folderID) == "" {
		fmt.Println("Error: missing folder ID.")
		printChangesHelp()
		return nil
	}

	plan, err := client.PlanFolder(folderID)
	if err != nil {
		return err
	}

	printPlan("PENDING CHANGES", plan)
	return nil
}

func runSync(args []string) error {
	folderID, err := resolveFolderID(args, "Folder ID")
	if err != nil {
		return err
	}
	if strings.TrimSpace(folderID) == "" {
		fmt.Println("Error: missing folder ID.")
		printSyncHelp()
		return nil
	}

	reportData, err := client.SyncFolder(folderID)
	if err != nil {
		return err
	}

	fmt.Println("------SYNC COMPLETE------")
	fmt.Printf("Added        : %d\n", reportData.Added)
	fmt.Printf("Modified     : %d\n", reportData.Modified)
	fmt.Printf("Deleted      : %d\n", reportData.Deleted)
	fmt.Printf("Bytes Copied : %d\n", reportData.BytesCopied)
	fmt.Printf("Duration     : %v\n", reportData.EndTime.Sub(reportData.StartTime))
	return nil
}

func runStatus(args []string) error {
	jsonOutput := hasFlag(args, "--json", "-j")

	folders, err := client.ListFolders()
	if err != nil {
		return err
	}

	if len(folders) == 0 {
		fmt.Println("No registered folders found.")
		return nil
	}

	if jsonOutput {
		data, err := json.MarshalIndent(folders, "", "  ")
		if err != nil {
			return err
		}
		fmt.Println(string(data))
		return nil
	}

	fmt.Println("ID | NAME | VERSION | LAST SYNC")
	fmt.Println("------------------------------------------------")
	for _, folder := range folders {
		fmt.Printf("%s | %s | %d | %s\n", folder.ID, folder.Name, folder.CurrentVersion, folder.LastSyncedAt.Format(time.RFC3339))
	}

	return nil
}

func runReport(args []string) error {
	folderID := ""
	if len(args) > 0 {
		folderID = strings.TrimSpace(args[0])
	} else {
		selected, err := resolveFolderID(args, "Folder ID")
		if err != nil {
			return err
		}
		folderID = selected
	}

	fileName, err := latestReportFile(folderID)
	if err != nil {
		return err
	}
	if fileName == "" {
		fmt.Println("Report not found.")
		return nil
	}

	data, err := os.ReadFile(fileName)
	if err != nil {
		return err
	}

	var reportData models.SyncReport
	if err := json.Unmarshal(data, &reportData); err != nil {
		return err
	}

	printSyncReport(fileName, reportData)
	return nil
}

func runServer(args []string) error {
	if len(args) == 0 || isHelpFlag(args[0]) {
		printServerHelp()
		return nil
	}

	switch args[0] {
	case "start":
		return serverStart()
	case "stop":
		return serverStop()
	case "status":
		return serverStatus(args[1:])
	case "logs":
		return serverLogs()
	default:
		printServerHelp()
		return nil
	}
}

func runSingleScan(folderPath string) error {
	if _, err := os.Stat(folderPath); err != nil {
		if os.IsNotExist(err) {
			fmt.Println("Folder path does not exist.")
			return nil
		}
		return err
	}

	result, err := scanFiles.Scanner(context.Background(), folderPath)
	if err != nil {
		return err
	}

	fmt.Printf("\n------SCAN------\n")
	fmt.Printf("Root       : %s\n", folderPath)
	fmt.Printf("Files      : %d\n", result.FilesCount)
	fmt.Printf("Directories: %d\n\n", result.DirCount)
	for _, file := range result.Files {
		fmt.Println(file)
	}
	return nil
}

func runTwoScan(sourceDir, destDir string) error {
	if err := runSingleScan(sourceDir); err != nil {
		return err
	}
	fmt.Println()
	return runSingleScan(destDir)
}

func runComparison(sourceDir, destDir string) error {
	result, srcMeta, destMeta, err := buildComparison(sourceDir, destDir)
	if err != nil {
		return err
	}

	fmt.Println("\n-----SOURCE-----")
	fmt.Printf("Root       : %s\n", sourceDir)
	fmt.Printf("Files      : %d\n", srcMeta.TotalFiles)
	fmt.Printf("Directories: %d\n", srcMeta.TotalDirs)
	fmt.Printf("Size       : %d bytes\n", srcMeta.TotalSize)

	fmt.Println("\n------DESTINATION-----")
	fmt.Printf("Root       : %s\n", destDir)
	fmt.Printf("Files      : %d\n", destMeta.TotalFiles)
	fmt.Printf("Directories: %d\n", destMeta.TotalDirs)
	fmt.Printf("Size       : %d bytes\n", destMeta.TotalSize)

	fmt.Println("\n-------COMPARISON-------")
	fmt.Printf("Added    : %d\n", result.Added)
	fmt.Printf("Modified : %d\n", result.Modified)
	fmt.Printf("Deleted  : %d\n", result.Deleted)
	fmt.Println("\n------CHANGES-------")

	if len(result.Changes) == 0 {
		fmt.Println("No changes detected.")
		return nil
	}

	for _, change := range result.Changes {
		fmt.Printf("[%s] %s\n", change.Type, change.RelativePath)
	}

	return nil
}

func promptOrArg(args []string, index int, label string) (string, error) {
	if len(args) > index {
		return strings.TrimSpace(args[index]), nil
	}

	return promptValue(label)
}

func resolveFolderID(args []string, label string) (string, error) {
	if len(args) > 0 {
		return strings.TrimSpace(args[0]), nil
	}

	folders, err := client.ListFolders()
	if err != nil {
		return "", err
	}
	if len(folders) == 0 {
		return "", nil
	}
	if len(folders) == 1 {
		return folders[0].ID, nil
	}

	fmt.Println("Select a folder:")
	for i, folder := range folders {
		fmt.Printf("  %d) %s | %s | %s\n", i+1, folder.ID, folder.Name, folder.LocalPath)
	}

	choice, err := promptValue(label + " number")
	if err != nil {
		return "", err
	}
	if choice == "" {
		return "", nil
	}

	index, err := strconv.Atoi(choice)
	if err != nil || index < 1 || index > len(folders) {
		fmt.Println("Invalid selection.")
		return "", nil
	}

	return folders[index-1].ID, nil
}

func hasFlag(args []string, flags ...string) bool {
	for _, arg := range args {
		for _, flag := range flags {
			if arg == flag {
				return true
			}
		}
	}
	return false
}

func promptValue(label string) (string, error) {
	fmt.Printf("%s: ", label)
	reader := bufio.NewReader(os.Stdin)
	value, err := reader.ReadString('\n')
	if err != nil && !errors.Is(err, io.EOF) {
		return "", err
	}

	return strings.TrimSpace(value), nil
}

func printPlan(title string, plan models.SyncPlan) {
	fmt.Println("------" + title + "------")
	fmt.Printf("Folder ID    : %s\n", plan.FolderID)
	fmt.Printf("Version      : %d\n", plan.Version)
	fmt.Printf("Uploads      : %d\n", len(plan.UploadPaths))
	fmt.Printf("Downloads    : %d\n", len(plan.DownloadFiles))
	fmt.Printf("Deletes      : %d\n", len(plan.DeletePaths))
	fmt.Printf("Conflicts    : %d\n", len(plan.Conflicts))
	for _, path := range plan.UploadPaths {
		fmt.Printf("[UPLOAD] %s\n", path)
	}
	for _, file := range plan.DownloadFiles {
		fmt.Printf("[DOWNLOAD] %s\n", file.RelativePath)
	}
	for _, path := range plan.DeletePaths {
		fmt.Printf("[DELETE] %s\n", path)
	}
	for _, conflict := range plan.Conflicts {
		fmt.Printf("[CONFLICT] %s -> %s\n", conflict.RelativePath, conflict.ConflictPath)
	}
}

func latestReportFile(folderID string) (string, error) {
	root := filepath.Join("storage", "client", "logs")
	if folderID != "" {
		path := filepath.Join(root, folderID+".json")
		if _, err := os.Stat(path); err == nil {
			return path, nil
		}
		return "", nil
	}

	entries, err := os.ReadDir(root)
	if err != nil {
		if os.IsNotExist(err) {
			return "", nil
		}
		return "", err
	}

	type fileInfo struct {
		path string
		mod  time.Time
	}

	files := make([]fileInfo, 0, len(entries))
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		info, err := entry.Info()
		if err != nil {
			continue
		}
		files = append(files, fileInfo{path: filepath.Join(root, entry.Name()), mod: info.ModTime()})
	}

	if len(files) == 0 {
		return "", nil
	}

	sort.Slice(files, func(i, j int) bool {
		return files[i].mod.After(files[j].mod)
	})

	return files[0].path, nil
}

func printSyncReport(fileName string, reportData models.SyncReport) {
	duration := reportData.EndTime.Sub(reportData.StartTime)
	fmt.Println("------SYNC REPORT------")
	fmt.Printf("File         : %s\n", fileName)
	fmt.Printf("Start Time   : %v\n", reportData.StartTime)
	fmt.Printf("End Time     : %v\n", reportData.EndTime)
	fmt.Printf("Duration     : %v\n", duration)
	fmt.Printf("Added        : %d\n", reportData.Added)
	fmt.Printf("Modified     : %d\n", reportData.Modified)
	fmt.Printf("Deleted      : %d\n", reportData.Deleted)
	fmt.Printf("Bytes Copied : %d\n", reportData.BytesCopied)
}

func serverStart() error {
	pidPath := filepath.Join("storage", "server", "config", "server.pid")
	if err := os.MkdirAll(filepath.Dir(pidPath), 0755); err != nil {
		return err
	}

	if err := os.WriteFile(pidPath, []byte(fmt.Sprintf("%d", os.Getpid())), 0644); err != nil {
		return err
	}

	app, err := serverapp.NewApp()
	if err != nil {
		return err
	}

	address := strings.TrimSpace(os.Getenv("FILESYNC_SERVER_ADDR"))
	if address == "" {
		address = ":8080"
	}
	fmt.Println("FileSync server starting on", address)
	return app.Listen(address)
}

func serverStop() error {
	pidPath := filepath.Join("storage", "server", "config", "server.pid")
	data, err := os.ReadFile(pidPath)
	if err != nil {
		fmt.Println("Server is not running.")
		return nil
	}

	pidString := strings.TrimSpace(string(data))
	if pidString == "" {
		fmt.Println("Server is not running.")
		return nil
	}

	pid, err := strconv.Atoi(pidString)
	if err != nil {
		return err
	}

	process, err := os.FindProcess(pid)
	if err != nil {
		return err
	}

	if err := process.Signal(os.Interrupt); err != nil {
		return err
	}

	return os.Remove(pidPath)
}

func serverStatus(args []string) error {
	jsonOutput := hasFlag(args, "--json", "-j")

	address := strings.TrimSpace(os.Getenv("FILESYNC_SERVER_ADDR"))
	if address == "" {
		address = "localhost:8080"
	} else if strings.HasPrefix(address, ":") {
		address = "localhost" + address
	}
	req, err := http.NewRequest(http.MethodGet, "http://"+address+"/status", nil)
	if err != nil {
		return err
	}
	if token := strings.TrimSpace(os.Getenv("FILESYNC_AUTH_TOKEN")); token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	resp, err := (&http.Client{Timeout: 10 * time.Second}).Do(req)
	if err != nil {
		fmt.Println("Connection failed.")
		return nil
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	if jsonOutput {
		var status models.ServerStatus
		if err := json.Unmarshal(body, &status); err != nil {
			return err
		}
		encoded, err := json.MarshalIndent(status, "", "  ")
		if err != nil {
			return err
		}
		fmt.Println(string(encoded))
		return nil
	}

	var status models.ServerStatus
	if err := json.Unmarshal(body, &status); err != nil {
		fmt.Println(string(body))
		return nil
	}

	fmt.Println("------SERVER STATUS------")
	fmt.Printf("Data Dir   : %s\n", status.DataDir)
	fmt.Printf("Blob Count : %d\n", status.BlobCount)
	fmt.Printf("Folders    : %d\n", len(status.Folders))
	for _, folder := range status.Folders {
		fmt.Printf("- %s | %s | v%d | %s\n", folder.ID, folder.Name, folder.Version, folder.UpdatedAt.Format(time.RFC3339))
	}
	return nil
}

func serverLogs() error {
	root := filepath.Join("storage", "server", "logs")
	entries, err := os.ReadDir(root)
	if err != nil {
		if os.IsNotExist(err) {
			fmt.Println("No server logs found.")
			return nil
		}
		return err
	}

	if len(entries) == 0 {
		fmt.Println("No server logs found.")
		return nil
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		data, err := os.ReadFile(filepath.Join(root, entry.Name()))
		if err != nil {
			continue
		}
		fmt.Printf("--- %s ---\n", entry.Name())
		fmt.Print(string(data))
	}

	return nil
}
