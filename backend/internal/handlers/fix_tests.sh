#!/bin/bash

# Add auth import to message_handler_test.go
sed -i '' '/import ($/a\n	"github.com/atom-yt/atom-ai-platform/backend/internal/auth"
' /Users/yangtong07/Desktop/code/llm-agent/claude-code-go/backend/internal/handlers/message_handler_test.go

# Update setupMessageTest to use auth.UserIDKey
sed -i '' 's/req = req.WithContext(setUserID(req.Context(), userID))/ctx := context.WithValue(req.Context(), auth.UserIDKey, userID)\n		req = req.WithContext(ctx)/g' /Users/yangtong07/Desktop/code/llm-agent/claude-code-go/backend/internal/handlers/message_handler_test.go

# Add auth import to agent_handler_test.go
sed -i '' '/import ($/a\n	"github.com/atom-yt/atom-ai-platform/backend/internal/auth"
' /Users/yangtong07/Desktop/code/llm-agent/claude-code-go/backend/internal/handlers/agent_handler_test.go

# Update setupAgentTest to use auth.UserIDKey
sed -i '' 's/req = req.WithContext(setUserID(req.Context(), userID))/ctx := context.WithValue(req.Context(), auth.UserIDKey, userID)\n		req = req.WithContext(ctx)/g' /Users/yangtong07/Desktop/code/llm-agent/claude-code-go/backend/internal/handlers/agent_handler_test.go

# Add auth import to skill_handler_test.go
sed -i '' '/import ($/a\n	"github.com/atom-yt/atom-ai-platform/backend/internal/auth"
' /Users/yangtong07/Desktop/code/llm-agent/claude-code-go/backend/internal/handlers/skill_handler_test.go

# Update setupSkillTest to use auth.UserIDKey
sed -i '' 's/req = req.WithContext(setUserID(req.Context(), userID))/ctx := context.WithValue(req.Context(), auth.UserIDKey, userID)\n		req = req.WithContext(ctx)/g' /Users/yangtong07/Desktop/code/llm-agent/claude-code-go/backend/internal/handlers/skill_handler_test.go

# Add auth import to artifact_handler_test.go
sed -i '' '/import ($/a\n	"github.com/atom-yt/atom-ai-platform/backend/internal/auth"
' /Users/yangtong07/Desktop/code/llm-agent/claude-code-go/backend/internal/handlers/artifact_handler_test.go

# Update setupArtifactTest to use auth.UserIDKey
sed -i '' 's/req = req.WithContext(setUserID(req.Context(), userID))/ctx := context.WithValue(req.Context(), auth.UserIDKey, userID)\n		req = req.WithContext(ctx)/g' /Users/yangtong07/Desktop/code/llm-agent/claude-code-go/backend/internal/handlers/artifact_handler_test.go

# Add auth import to schedule_handler_test.go
sed -i '' '/import ($/a\n	"github.com/atom-yt/atom-ai-platform/backend/internal/auth"
' /Users/yangtong07/Desktop/code/llm-agent/claude-code-go/backend/internal/handlers/schedule_handler_test.go

# Update setupScheduleTest to use auth.UserIDKey
sed -i '' 's/req = req.WithContext(setUserID(req.Context(), userID))/ctx := context.WithValue(req.Context(), auth.UserIDKey, userID)\n		req = req.WithContext(ctx)/g' /Users/yangtong07/Desktop/code/llm-agent/claude-code-go/backend/internal/handlers/schedule_handler_test.go

# Add auth import to knowledge_handler_test.go
sed -i '' '/import ($/a\n	"github.com/atom-yt/atom-ai-platform/backend/internal/auth"
' /Users/yangtong07/Desktop/code/llm-agent/claude-code-go/backend/internal/handlers/knowledge_handler_test.go

# Update setupKnowledgeTest to use auth.UserIDKey
sed -i '' 's/req = req.WithContext(setUserID(req.Context(), userID))/ctx := context.WithValue(req.Context(), auth.UserIDKey, userID)\n		req = req.WithContext(ctx)/g' /Users/yangtong07/Desktop/code/llm-agent/claude-code-go/backend/internal/handlers/knowledge_handler_test.go

# Add auth import to chat_handler_test.go
sed -i '' '/import ($/a\n	"github.com/atom-yt/atom-ai-platform/backend/internal/auth"
' /Users/yangtong07/Desktop/code/llm-agent/claude-code-go/backend/internal/handlers/chat_handler_test.go

# Update chat_handler_test to use auth.UserIDKey
sed -i '' 's/req = req.WithContext(setUserID(req.Context(), userID))/ctx := context.WithValue(req.Context(), auth.UserIDKey, userID)\n		req = req.WithContext(ctx)/g' /Users/yangtong07/Desktop/code/llm-agent/claude-code-go/backend/internal/handlers/chat_handler_test.go

echo "Test files updated"
