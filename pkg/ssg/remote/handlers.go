// Package remote provides RPC handlers for SSG Git operations.
package remote

import (
	"context"
	"fmt"

	"github.com/cyw0ng95/v2e/pkg/proc/subprocess"
)

// RegisterHandlers registers all SSG remote RPC handlers.
func RegisterHandlers(sp *subprocess.Subprocess, gitClient *GitClient) {
	sp.RegisterHandler("RPCSSGCloneRepo", createCloneRepoHandler(gitClient))
	sp.RegisterHandler("RPCSSGPullRepo", createPullRepoHandler(gitClient))
	sp.RegisterHandler("RPCSSGGetRepoStatus", createGetRepoStatusHandler(gitClient))
	sp.RegisterHandler("RPCSSGListGuideFiles", createListGuideFilesHandler(gitClient))
	sp.RegisterHandler("RPCSSGListTableFiles", createListTableFilesHandler(gitClient))
	sp.RegisterHandler("RPCSSGListManifestFiles", createListManifestFilesHandler(gitClient))
	sp.RegisterHandler("RPCSSGGetFilePath", createGetFilePathHandler(gitClient))
}

// createCloneRepoHandler creates a handler for RPCSSGCloneRepo.
func createCloneRepoHandler(gitClient *GitClient) subprocess.Handler {
	return func(ctx context.Context, msg *subprocess.Message) (*subprocess.Message, error) {
		if err := gitClient.Clone(); err != nil {
			return subprocess.NewErrorResponse(msg, fmt.Sprintf("failed to clone repository: %s", err)), nil
		}

		return subprocess.NewSuccessResponse(msg, map[string]interface{}{
			"success": true,
			"path":    gitClient.GetFilePath(""),
		})
	}
}

// createPullRepoHandler creates a handler for RPCSSGPullRepo.
func createPullRepoHandler(gitClient *GitClient) subprocess.Handler {
	return func(ctx context.Context, msg *subprocess.Message) (*subprocess.Message, error) {
		if err := gitClient.Pull(); err != nil {
			return subprocess.NewErrorResponse(msg, fmt.Sprintf("failed to pull repository: %s", err)), nil
		}

		return subprocess.NewSuccessResponse(msg, map[string]interface{}{
			"success": true,
		})
	}
}

// createGetRepoStatusHandler creates a handler for RPCSSGGetRepoStatus.
func createGetRepoStatusHandler(gitClient *GitClient) subprocess.Handler {
	return func(ctx context.Context, msg *subprocess.Message) (*subprocess.Message, error) {
		status, err := gitClient.Status()
		if err != nil {
			return subprocess.NewErrorResponse(msg, fmt.Sprintf("failed to get repository status: %s", err)), nil
		}

		return subprocess.NewSuccessResponse(msg, map[string]interface{}{
			"commit_hash": status.CommitHash,
			"branch":      status.Branch,
			"is_clean":    status.IsClean,
		})
	}
}

// createListGuideFilesHandler creates a handler for RPCSSGListGuideFiles.
func createListGuideFilesHandler(gitClient *GitClient) subprocess.Handler {
	return func(ctx context.Context, msg *subprocess.Message) (*subprocess.Message, error) {
		files, err := gitClient.ListGuideFiles()
		if err != nil {
			return subprocess.NewErrorResponse(msg, fmt.Sprintf("failed to list guide files: %s", err)), nil
		}

		return subprocess.NewSuccessResponse(msg, map[string]interface{}{
			"files": files,
			"count": len(files),
		})
	}
}

// createListTableFilesHandler creates a handler for RPCSSGListTableFiles.
func createListTableFilesHandler(gitClient *GitClient) subprocess.Handler {
	return func(ctx context.Context, msg *subprocess.Message) (*subprocess.Message, error) {
		files, err := gitClient.ListTableFiles()
		if err != nil {
			return subprocess.NewErrorResponse(msg, fmt.Sprintf("failed to list table files: %s", err)), nil
		}

		return subprocess.NewSuccessResponse(msg, map[string]interface{}{
			"files": files,
			"count": len(files),
		})
	}
}

// createListManifestFilesHandler creates a handler for RPCSSGListManifestFiles.
func createListManifestFilesHandler(gitClient *GitClient) subprocess.Handler {
	return func(ctx context.Context, msg *subprocess.Message) (*subprocess.Message, error) {
		files, err := gitClient.ListManifestFiles()
		if err != nil {
			return subprocess.NewErrorResponse(msg, fmt.Sprintf("failed to list manifest files: %s", err)), nil
		}

		return subprocess.NewSuccessResponse(msg, map[string]interface{}{
			"files": files,
			"count": len(files),
		})
	}
}

// createGetFilePathHandler creates a handler for RPCSSGGetFilePath.
func createGetFilePathHandler(gitClient *GitClient) subprocess.Handler {
	return func(ctx context.Context, msg *subprocess.Message) (*subprocess.Message, error) {
		var req struct {
			Filename string `json:"filename"`
		}

		if msg.Payload != nil {
			_ = subprocess.UnmarshalPayload(msg, &req)
		}

		if errMsg := subprocess.RequireField(msg, req.Filename, "filename"); errMsg != nil {
			return errMsg, nil
		}

		path := gitClient.GetFilePath(req.Filename)

		return subprocess.NewSuccessResponse(msg, map[string]interface{}{
			"path": path,
		})
	}
}
