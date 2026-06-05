package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strconv"

	"FileSync/internal/comparator"
	"FileSync/internal/metadata"
	"FileSync/internal/models"
	"FileSync/internal/pairmanager"
	"FileSync/internal/report"
	"FileSync/internal/scanFiles"
	"FileSync/internal/synchronizer"
)

func buildComparison(sourceDir string, destDir string) (models.CompareResult, models.DirMetaData, models.DirMetaData, error) {
	srcScan, err := scanFiles.Scanner(context.Background(), sourceDir)
	if err != nil {
		return models.CompareResult{}, models.DirMetaData{}, models.DirMetaData{}, err
	}

	destScan, err := scanFiles.Scanner(context.Background(), destDir)
	if err != nil {
		return models.CompareResult{}, models.DirMetaData{}, models.DirMetaData{}, err
	}


	srcFiles, srcMeta, err := metadata.Build(srcScan)
	if err != nil {
		return models.CompareResult{}, models.DirMetaData{}, models.DirMetaData{}, err
	}

	destFiles, destMeta, err := metadata.Build(destScan)
	if err != nil {
		return models.CompareResult{}, models.DirMetaData{}, models.DirMetaData{}, err
	}

	result := comparator.Compare(srcFiles, destFiles)

	return result, srcMeta, destMeta, nil
}

func RunAdd() error {
	if len(os.Args) != 5 {
		fmt.Println("Usage: fs add <source> <destination>")
		return nil
	}

	pair, err := pairmanager.Add(os.Args[3], os.Args[4])
	if err != nil {
		return err
	}

	fmt.Printf("Pair Created: %d\n", pair.ID)

	return nil
}

func RunList() error {
	pairs, err := pairmanager.List()
	if err != nil {
		return err
	}

	fmt.Println("ID | SOURCE | DESTINATION")
	fmt.Println("----------------------------")
        

	for _, pair := range pairs {
		fmt.Printf("%d | %s -> %s\n",
			pair.ID,
			pair.Source,
			pair.Destination,
		)
	}

	return nil
}

func RunDetails() error {
	if len(os.Args) != 4 {
		fmt.Println("Usage: fs details <pair-id>")
		return nil
	}

	id, err := strconv.Atoi(os.Args[3])
	if err != nil {
		return err
	}

	pair, err := pairmanager.Get(id)
	if err != nil {
		return err
	}

	if pair == nil {
		fmt.Println("Pair not found.")
		return nil
	}

	fmt.Println("------PAIR DETAILS------")
	fmt.Printf("ID          : %d\n", pair.ID)
	fmt.Printf("Source      : %s\n", pair.Source)
	fmt.Printf("Destination : %s\n", pair.Destination)
	fmt.Printf("Created     : %s\n", pair.CreatedAt)

	return nil
}

func RunDelete() error {
	if len(os.Args) != 4 {
		fmt.Println("Usage: fs delete <pair-id>")
		return nil
	}

	id, err := strconv.Atoi(os.Args[3])
	if err != nil {
		return err
	}

	err = pairmanager.Delete(id)
	if err != nil {
		return err
	}

	fmt.Printf("Pair %d deleted.\n", id)

	return nil
}

func RunCompare() error {
	if len(os.Args) != 5 {
		fmt.Println("Usage: fs compare <source> <destination>")
		return nil
	}

	result, srcMeta, destMeta, err := buildComparison(os.Args[3], os.Args[4])
	if err != nil {
		return err
	}

	fmt.Println("\n-----SOURCE-----")
	fmt.Printf("Root       : %s\n", os.Args[3])
	fmt.Printf("Files      : %d\n", srcMeta.TotalFiles)
	fmt.Printf("Directories: %d\n", srcMeta.TotalDirs)
	fmt.Printf("Size       : %d bytes\n", srcMeta.TotalSize)

	fmt.Println("\n------DESTINATION-----")
	fmt.Printf("Root       : %s\n", os.Args[4])
	fmt.Printf("Files      : %d\n", destMeta.TotalFiles)
	fmt.Printf("Directories: %d\n", destMeta.TotalDirs)
	fmt.Printf("Size       : %d bytes\n", destMeta.TotalSize)

	fmt.Println("\n-------COMPARISON-------")
	fmt.Printf("Added    : %d\n", result.Added)
	fmt.Printf("Modified : %d\n", result.Modified)
	fmt.Printf("Deleted  : %d\n", result.Deleted)

	fmt.Println("\n------CHANGES-------")

	for _, change := range result.Changes {
		fmt.Printf("[%s] %s\n", change.Type, change.RelativePath)
	}

	return nil
}

func RunChanges() error {
	if len(os.Args) != 4 {
		fmt.Println("Usage: fs changes <pair-id>")
		return nil
	}

	id, _ := strconv.Atoi(os.Args[3])

	pair, err := pairmanager.Get(id)
	if err != nil {
		return err
	}

	if pair == nil {
		fmt.Println("Pair not found.")
		return nil
	}

	result, _, _, err := buildComparison(pair.Source, pair.Destination)
	if err != nil {
		return err
	}

	fmt.Println("------CHANGES------")

	for _, change := range result.Changes {
		fmt.Printf("[%s] %s\n", change.Type, change.RelativePath)
	}

	return nil
}

func RunSync() error {
	if len(os.Args) != 4 {
		fmt.Println("Usage: fs sync <pair-id>")
		return nil
	}

	id, _ := strconv.Atoi(os.Args[3])

	pair, err := pairmanager.Get(id)
	if err != nil {
		return err
	}

	if pair == nil {
		fmt.Println("Pair not found.")
		return nil
	}

	result, _, _, err := buildComparison(pair.Source, pair.Destination)
	if err != nil {
		return err
	}

	syncReport, err := synchronizer.Sync(pair.Source, pair.Destination, result)
	if err != nil {
		return err
	}

	err = report.Save(pair.ID, syncReport)
	if err != nil {
		return err
	}

	duration := syncReport.EndTime.Sub(syncReport.StartTime)

	fmt.Println("------SYNC COMPLETE------")
	fmt.Printf("Added        : %d\n", syncReport.Added)
	fmt.Printf("Modified     : %d\n", syncReport.Modified)
	fmt.Printf("Deleted      : %d\n", syncReport.Deleted)
	fmt.Printf("Bytes Copied : %d\n", syncReport.BytesCopied)
	fmt.Printf("Duration     : %v\n", duration)

	return nil
}

func RunReport() error {
	if len(os.Args) != 4 {
		fmt.Println("Usage: fs report <pair-id>")
		return nil
	}
	id, err := strconv.Atoi(os.Args[3])
	if err != nil {
		fmt.Println("Invalid pair ID.")
		return nil
	}

	file := fmt.Sprintf("reports/pair-%d.json", id)

	if _, err := os.Stat(file); os.IsNotExist(err) {
		fmt.Println("Report not found.")
		return nil
	}

	data, err := os.ReadFile(file)
	if err != nil {
		return err
	}

	var reportData models.SyncReport

	err = json.Unmarshal(data, &reportData)
	if err != nil {
		return err
	}

	fmt.Println("------SYNC REPORT------")

	fmt.Printf("Added        : %d\n", reportData.Added)
	fmt.Printf("Modified     : %d\n", reportData.Modified)
	fmt.Printf("Deleted      : %d\n", reportData.Deleted)
	fmt.Printf("Bytes Copied : %d\n", reportData.BytesCopied)

	fmt.Println("\n------Added Files------")
	for _, file := range reportData.AddedFiles {
		fmt.Println(file)
	}

	fmt.Println("\n-------Modified Files-------")
	for _, file := range reportData.ModifiedFiles {
		fmt.Println(file)
	}

	fmt.Println("\n-------Deleted Files--------")
	for _, file := range reportData.DeletedFiles {
		fmt.Println(file)
	}

	return nil
}

func RunScanner() error {
	if len(os.Args) != 5 {
		fmt.Println("Usage: go run main.go fs scan <source_dir> <dest_dir>")
		return nil
	}

	sourceDir := os.Args[3]
	destDir := os.Args[4]

	srcScan, err := scanFiles.Scanner(
		context.Background(),
		sourceDir,
	)
	if err != nil {
		return err
	}

	fmt.Printf("\n-----SOURCE-----\n")
	fmt.Printf("Root       : %s\n", sourceDir)
	fmt.Printf("Files      : %d\n", srcScan.FilesCount)
	fmt.Printf("Directories: %d\n\n", srcScan.DirCount)

	for _, file := range srcScan.Files {
		fmt.Println(file)
	}

	destScan, err := scanFiles.Scanner(
		context.Background(),
		destDir,
	)
	if err != nil {
		return err
	}
	fmt.Printf("\n------DESTINATION------\n")
	fmt.Printf("Root       : %s\n", sourceDir)
	fmt.Printf("Files      : %d\n", srcScan.FilesCount)
	fmt.Printf("Directories: %d\n\n", srcScan.DirCount)

	for _, file := range destScan.Files {
		fmt.Println(file)
	}

	return nil
}
