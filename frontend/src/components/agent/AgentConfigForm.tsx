import React, { useState, useEffect } from "react";
import { AgentConfig, AgentConfigFormProps, FormErrors } from "./types";
import { MODEL_PRESETS, getModelById, getProviderById, validateModelProviderPair } from "./constants";
import { ModelSelect } from "./ModelSelect";
import { ProviderSelect } from "./ProviderSelect";
import { Input } from "@/components/ui/input";
import { Textarea } from "@/components/ui/textarea";
import { Button } from "@/components/ui/button";
import { Label } from "@/components/ui/label";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { Badge } from "@/components/ui/badge";
import { Loader2, Save, X, CheckCircle2, AlertCircle } from "lucide-react";

export const AgentConfigForm: React.FC<AgentConfigFormProps> = ({
  initialConfig,
  onSave,
  onCancel,
  isLoading = false,
}) => {
  const [config, setConfig] = useState<AgentConfig>(initialConfig || {});
  const [errors, setErrors] = useState<FormErrors>({});
  const [saveStatus, setSaveStatus] = useState<"idle" | "success" | "error">("idle");
  const [errorMessage, setErrorMessage] = useState("");

  // Update form when initialConfig changes
  useEffect(() => {
    if (initialConfig) {
      setConfig(initialConfig);
    }
  }, [initialConfig]);

  const validate = (): boolean => {
    const newErrors: FormErrors = {};

    if (!config.provider) {
      newErrors.provider = "Provider is required";
    }

    if (!config.model) {
      newErrors.model = "Model is required";
    } else if (config.provider && !validateModelProviderPair(config.model, config.provider)) {
      newErrors.model = "Model does not match selected provider";
    }

    // Custom API key validation
    if (config.api_key && config.api_key.length < 8) {
      newErrors.api_key = "API key must be at least 8 characters";
    }

    // Base URL validation
    if (config.base_url) {
      try {
        new URL(config.base_url);
      } catch {
        newErrors.base_url = "Invalid URL format";
      }
    }

    setErrors(newErrors);
    return Object.keys(newErrors).length === 0;
  };

  const handleSave = async () => {
    if (!validate()) {
      return;
    }

    try {
      await onSave(config);
      setSaveStatus("success");
      setErrorMessage("");
      // Reset status after 3 seconds
      setTimeout(() => setSaveStatus("idle"), 3000);
    } catch (error: any) {
      setSaveStatus("error");
      setErrorMessage(error.message || "Failed to save configuration");
    }
  };

  const handlePresetSelect = (preset: typeof MODEL_PRESETS[number]) => {
    setConfig({
      ...config,
      model: preset.config.model,
      provider: preset.config.provider,
    });
    setErrors((prev) => ({
      ...prev,
      model: undefined,
      provider: undefined,
    }));
  };

  const handleFieldChange = (field: keyof AgentConfig, value: string) => {
    setConfig((prev) => ({ ...prev, [field]: value }));
    setErrors((prev) => ({ ...prev, [field]: undefined }));

    // Auto-select provider when model changes
    if (field === "model" && value) {
      const model = getModelById(value);
      if (model && model.provider !== config.provider) {
        setConfig((prev) => ({ ...prev, provider: model.provider }));
      }
    }
  };

  const resetForm = () => {
    setConfig(initialConfig || {});
    setErrors({});
    setSaveStatus("idle");
    setErrorMessage("");
  };

  const selectedModel = config.model ? getModelById(config.model) : null;
  const selectedProvider = config.provider ? getProviderById(config.provider) : null;

  return (
    <div className="space-y-6">
      {/* Presets Section */}
      <div className="space-y-3">
        <Label>Quick Presets</Label>
        <div className="flex flex-wrap gap-2">
          {MODEL_PRESETS.map((preset) => (
            <Badge
              key={preset.id}
              variant={
                config.model === preset.config.model &&
                config.provider === preset.config.provider
                  ? "default"
                  : "outline"
              }
              className="cursor-pointer hover:bg-primary/80"
              onClick={() => handlePresetSelect(preset)}
            >
              {preset.name}
            </Badge>
          ))}
        </div>
      </div>

      {/* Form Fields */}
      <div className="space-y-4">
        <ProviderSelect
          value={config.provider}
          onChange={(value) => handleFieldChange("provider", value)}
          disabled={isLoading}
        />

        <ModelSelect
          value={config.model}
          onChange={(value) => handleFieldChange("model", value)}
          provider={config.provider}
          disabled={isLoading}
        />

        <Tabs defaultValue="manual" className="w-full">
          <TabsList className="grid w-full grid-cols-2">
            <TabsTrigger value="manual">Manual Configuration</TabsTrigger>
            <TabsTrigger value="advanced">Advanced</TabsTrigger>
          </TabsList>

          <TabsContent value="manual" className="space-y-4">
            <div className="space-y-2">
              <Label htmlFor="api_key">
                API Key
                {selectedProvider && (
                  <span className="text-xs text-muted-foreground ml-2">
                    ({selectedProvider.envKey})
                  </span>
                )}
              </Label>
              <Input
                id="api_key"
                type="password"
                placeholder="Enter API key (optional if set via environment)"
                value={config.api_key || ""}
                onChange={(e) => handleFieldChange("api_key", e.target.value)}
                disabled={isLoading}
                className={errors.api_key ? "border-destructive" : ""}
              />
              {errors.api_key && (
                <p className="text-sm text-destructive flex items-center gap-1">
                  <AlertCircle className="w-4 h-4" />
                  {errors.api_key}
                </p>
              )}
            </div>

            <div className="space-y-2">
              <Label htmlFor="base_url">Base URL</Label>
              <Input
                id="base_url"
                type="url"
                placeholder={selectedProvider?.baseURL || "https://api.example.com"}
                value={config.base_url || ""}
                onChange={(e) => handleFieldChange("base_url", e.target.value)}
                disabled={isLoading}
                className={errors.base_url ? "border-destructive" : ""}
              />
              {errors.base_url && (
                <p className="text-sm text-destructive flex items-center gap-1">
                  <AlertCircle className="w-4 h-4" />
                  {errors.base_url}
                </p>
              )}
            </div>
          </TabsContent>

          <TabsContent value="advanced" className="space-y-4">
            <div className="space-y-2">
              <Label htmlFor="system_prompt">System Prompt</Label>
              <Textarea
                id="system_prompt"
                placeholder="You are a helpful AI assistant..."
                value={config.system_prompt || ""}
                onChange={(e) => handleFieldChange("system_prompt", e.target.value)}
                disabled={isLoading}
                rows={6}
                className={errors.system_prompt ? "border-destructive" : ""}
              />
              {errors.system_prompt && (
                <p className="text-sm text-destructive flex items-center gap-1">
                  <AlertCircle className="w-4 h-4" />
                  {errors.system_prompt}
                </p>
              )}
            </div>

            {/* Config Summary */}
            <div className="p-4 rounded-md border bg-muted/50 space-y-2">
              <h4 className="font-medium text-sm">Configuration Summary</h4>
              <div className="grid grid-cols-2 gap-2 text-xs">
                <div className="flex flex-col">
                  <span className="text-muted-foreground">Provider:</span>
                  <span className="font-medium">
                    {selectedProvider?.name || "Not selected"}
                  </span>
                </div>
                <div className="flex flex-col">
                  <span className="text-muted-foreground">Model:</span>
                  <span className="font-medium">
                    {selectedModel?.name || "Not selected"}
                  </span>
                </div>
                <div className="flex flex-col">
                  <span className="text-muted-foreground">Tool Use:</span>
                  <span className="font-medium">
                    {selectedModel?.capabilities?.toolUse ? "Yes" : "No"}
                  </span>
                </div>
                <div className="flex flex-col">
                  <span className="text-muted-foreground">Vision:</span>
                  <span className="font-medium">
                    {selectedModel?.capabilities?.vision ? "Yes" : "No"}
                  </span>
                </div>
                {selectedModel?.contextWindow && (
                  <div className="flex flex-col col-span-2">
                    <span className="text-muted-foreground">Context Window:</span>
                    <span className="font-medium">
                      {selectedModel.contextWindow.toLocaleString()} tokens
                    </span>
                  </div>
                )}
              </div>
            </div>
          </TabsContent>
        </Tabs>
      </div>

      {/* Actions */}
      <div className="flex items-center justify-between gap-3 pt-4 border-t">
        <Button
          type="button"
          variant="outline"
          onClick={onCancel}
          disabled={isLoading}
        >
          <X className="w-4 h-4 mr-2" />
          Cancel
        </Button>

        <div className="flex items-center gap-2">
          {saveStatus === "success" && (
            <div className="text-sm text-green-600 flex items-center gap-1">
              <CheckCircle2 className="w-4 h-4" />
              Saved successfully
            </div>
          )}
          {saveStatus === "error" && (
            <div className="text-sm text-destructive flex items-center gap-1">
              <AlertCircle className="w-4 h-4" />
              {errorMessage}
            </div>
          )}
          <Button
            type="button"
            onClick={handleSave}
            disabled={isLoading}
            className="min-w-[100px]"
          >
            {isLoading ? (
              <>
                <Loader2 className="w-4 h-4 mr-2 animate-spin" />
                Saving...
              </>
            ) : (
              <>
                <Save className="w-4 h-4 mr-2" />
                Save
              </>
            )}
          </Button>
        </div>
      </div>
    </div>
  );
};
