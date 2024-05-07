package main

import (
	"fmt"

	"github.com/urfave/cli/v2"

	"github.com/Layr-Labs/op-plasma-eigenda/eigenda"
	opservice "github.com/ethereum-optimism/optimism/op-service"
	oplog "github.com/ethereum-optimism/optimism/op-service/log"
)

const (
	ListenAddrFlagName    = "addr"
	PortFlagName          = "port"
	S3BucketFlagName      = "s3.bucket"
	FileStorePathFlagName = "file.path"
)

const EnvVarPrefix = "OP_PLASMA_DA_SERVER"

func prefixEnvVars(name string) []string {
	return opservice.PrefixEnvVar(EnvVarPrefix, name)
}

var (
	ListenAddrFlag = &cli.StringFlag{
		Name:    ListenAddrFlagName,
		Usage:   "server listening address",
		Value:   "127.0.0.1",
		EnvVars: prefixEnvVars("ADDR"),
	}
	PortFlag = &cli.IntFlag{
		Name:    PortFlagName,
		Usage:   "server listening port",
		Value:   3100,
		EnvVars: prefixEnvVars("PORT"),
	}
	FileStorePathFlag = &cli.StringFlag{
		Name:    FileStorePathFlagName,
		Usage:   "path to directory for file storage",
		EnvVars: prefixEnvVars("FILESTORE_PATH"),
	}
	S3BucketFlag = &cli.StringFlag{
		Name:    S3BucketFlagName,
		Usage:   "bucket name for S3 storage",
		EnvVars: prefixEnvVars("S3_BUCKET"),
	}
)

var requiredFlags = []cli.Flag{
	ListenAddrFlag,
	PortFlag,
}

var optionalFlags = []cli.Flag{
	FileStorePathFlag,
	S3BucketFlag,
}

func init() {
	optionalFlags = append(optionalFlags, oplog.CLIFlags(EnvVarPrefix)...)
	optionalFlags = append(optionalFlags, eigenda.CLIFlags(EnvVarPrefix)...)
	Flags = append(requiredFlags, optionalFlags...)
}

// Flags contains the list of configuration options available to the binary.
var Flags []cli.Flag

type CLIConfig struct {
	FileStoreDirPath string
	S3Bucket         string
	EigenDAConfig    eigenda.Config
}

func ReadCLIConfig(ctx *cli.Context) CLIConfig {
	return CLIConfig{
		FileStoreDirPath: ctx.String(FileStorePathFlagName),
		S3Bucket:         ctx.String(S3BucketFlagName),
		EigenDAConfig:    eigenda.ReadConfig(ctx),
	}
}

func (c CLIConfig) Check() error {
	enabledStores := 0
	if c.S3Enabled() {
		enabledStores += 1
	}
	if c.FileStoreEnabled() {
		enabledStores += 1
	}
	if c.EigenDAEnabled() {
		err := c.EigenDAConfig.Check()
		if err != nil {
			return err
		}
		enabledStores += 1
	}
	if enabledStores == 0 {
		return fmt.Errorf("at least one storage backend must be enabled")
	}
	if enabledStores > 1 {
		return fmt.Errorf("only one storage backend can be enabled")
	}
	return nil
}

func (c CLIConfig) S3Enabled() bool {
	return c.S3Bucket != ""
}

func (c CLIConfig) FileStoreEnabled() bool {
	return c.FileStoreDirPath != ""
}

func (c CLIConfig) EigenDAEnabled() bool {
	return c.EigenDAConfig.RPC != ""
}

func CheckRequired(ctx *cli.Context) error {
	for _, f := range requiredFlags {
		if !ctx.IsSet(f.Names()[0]) {
			return fmt.Errorf("flag %s is required", f.Names()[0])
		}
	}
	return nil
}
