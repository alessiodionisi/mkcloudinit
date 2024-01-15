package main

import (
	"errors"
	"fmt"
	"os"

	"github.com/diskfs/go-diskfs"
	"github.com/diskfs/go-diskfs/disk"
	"github.com/diskfs/go-diskfs/filesystem"
	"github.com/diskfs/go-diskfs/filesystem/iso9660"
	"github.com/spf13/cobra"
)

type commandOptions struct {
	metaData      string
	networkConfig string
	output        string
	userData      string
}

func main() {
	if err := newCommand().Execute(); err != nil {
		os.Exit(1)
	}
}

func newCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "mkcloudinit",
		Short: "Build an iso9660 cloud-init image in seconds",
		RunE: func(cmd *cobra.Command, _ []string) error {
			output, err := cmd.Flags().GetString("output")
			if err != nil {
				return err
			}

			userData, err := cmd.Flags().GetString("user-data")
			if err != nil {
				return err
			}

			metaData, err := cmd.Flags().GetString("meta-data")
			if err != nil {
				return err
			}

			networkConfig, err := cmd.Flags().GetString("network-config")
			if err != nil {
				return err
			}

			opts := &commandOptions{
				metaData:      metaData,
				networkConfig: networkConfig,
				output:        output,
				userData:      userData,
			}

			if err := runCommand(opts); err != nil {
				fmt.Printf("%s", err)
				os.Exit(1)
			}

			return nil
		},
	}

	cmd.Flags().StringP("meta-data", "m", "", "meta-data content")
	cmd.Flags().StringP("network-config", "n", "", "network-config content")
	cmd.Flags().StringP("output", "o", "", "output file name")
	cmd.Flags().StringP("user-data", "u", "", "user-data content")

	if err := cmd.MarkFlagRequired("output"); err != nil {
		fmt.Printf("%s", err)
		os.Exit(1)
	}

	return cmd
}

func runCommand(opts *commandOptions) error {
	size := 10 * 1024 * 1024

	diskFile, err := diskfs.Create(opts.output, int64(size), diskfs.Raw, diskfs.SectorSizeDefault)
	if err != nil {
		return err
	}

	diskFile.LogicalBlocksize = 2048

	diskFileSystem, err := diskFile.CreateFilesystem(disk.FilesystemSpec{
		Partition: 0,
		FSType:    filesystem.TypeISO9660,
	})
	if err != nil {
		return err
	}

	// meta data
	if opts.metaData != "" {
		metaDataFile, err := diskFileSystem.OpenFile("meta-data", os.O_CREATE|os.O_RDWR)
		if err != nil {
			return err
		}
		defer metaDataFile.Close()

		if _, err := metaDataFile.Write([]byte(opts.metaData)); err != nil {
			return err
		}
	}

	// user data
	if opts.userData != "" {
		userDataFile, err := diskFileSystem.OpenFile("user-data", os.O_CREATE|os.O_RDWR)
		if err != nil {
			return err
		}
		defer userDataFile.Close()

		if _, err := userDataFile.Write([]byte(opts.userData)); err != nil {
			return err
		}
	}

	// network config
	if opts.networkConfig != "" {
		networkConfigFile, err := diskFileSystem.OpenFile("network-config", os.O_CREATE|os.O_RDWR)
		if err != nil {
			return err
		}
		defer networkConfigFile.Close()

		if _, err := networkConfigFile.Write([]byte(opts.networkConfig)); err != nil {
			return err
		}
	}

	isoFileSystem, ok := diskFileSystem.(*iso9660.FileSystem)
	if !ok {
		return errors.New("invalid file system")
	}

	return isoFileSystem.Finalize(iso9660.FinalizeOptions{
		RockRidge:        true,
		VolumeIdentifier: "cidata",
	})
}
