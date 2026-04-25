import React, { useState, useMemo } from "react";
import { ModelSelectProps } from "./types";
import { MODELS, PROVIDERS, getModelById, getModelsByProvider } from "./constants";
import {
  Select,
  SelectContent,
  SelectGroup,
  SelectItem,
  SelectLabel,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { Label } from "@/components/ui/label";
import { Input } from "@/components/ui/input";
import { Search, Cpu, Database } from "lucide-react";

export const ModelSelect: React.FC<ModelSelectProps> = ({
  value,
  onChange,
  provider,
  disabled = false,
}) => {
  const [searchTerm, setSearchTerm] = useState("");
  const [isOpen, setIsOpen] = useState(false);

  const availableProviders = useMemo(() => {
    if (!provider) return PROVIDERS;
    return PROVIDERS.filter((p) => p.id === provider);
  }, [provider]);

  const filteredModels = useMemo(() => {
    let models = provider ? getModelsByProvider(provider) : MODELS;

    if (searchTerm.trim()) {
      const term = searchTerm.toLowerCase();
      models = models.filter(
        (m) =>
          m.name.toLowerCase().includes(term) ||
          m.description.toLowerCase().includes(term) ||
          m.provider.toLowerCase().includes(term)
      );
    }

    return models;
  }, [provider, searchTerm]);

  const selectedModel = useMemo(() => {
    return value ? getModelById(value) : null;
  }, [value]);

  const groupedModels = useMemo(() => {
    const groups: Record<string, typeof MODELS> = {};
    availableProviders.forEach((p) => {
      const providerModels = filteredModels.filter((m) => m.provider === p.id);
      if (providerModels.length > 0) {
        groups[p.name] = providerModels;
      }
    });
    return groups;
  }, [filteredModels, availableProviders]);

  return (
    <div className="space-y-2">
      <Label htmlFor="model">Model</Label>
      <Select
        value={value}
        onValueChange={onChange}
        disabled={disabled}
        open={isOpen}
        onOpenChange={setIsOpen}
      >
        <SelectTrigger id="model">
          <SelectValue placeholder="Select a model" />
        </SelectTrigger>
        <SelectContent>
          {isOpen && (
            <div className="p-2">
              <div className="relative">
                <Search className="absolute left-2 top-2.5 h-4 w-4 text-muted-foreground" />
                <Input
                  placeholder="Search models..."
                  value={searchTerm}
                  onChange={(e) => setSearchTerm(e.target.value)}
                  className="pl-8"
                  onClick={(e) => e.stopPropagation()}
                />
              </div>
            </div>
          )}
          {Object.keys(groupedModels).length === 0 ? (
            <div className="p-4 text-center text-sm text-muted-foreground">
              No models found
            </div>
          ) : (
            Object.entries(groupedModels).map(([providerName, models]) => (
              <SelectGroup key={providerName}>
                {Object.keys(groupedModels).length > 1 && (
                  <SelectLabel>{providerName}</SelectLabel>
                )}
                {models.map((model) => (
                  <SelectItem key={model.id} value={model.id}>
                    <div className="flex flex-col gap-1">
                      <div className="flex items-center justify-between w-full">
                        <span className="font-medium">{model.name}</span>
                        {model.capabilities?.vision && (
                          <span className="text-xs text-muted-foreground">Vision</span>
                        )}
                      </div>
                      <span className="text-xs text-muted-foreground line-clamp-1">
                        {model.description}
                      </span>
                    </div>
                  </SelectItem>
                ))}
              </SelectGroup>
            ))
          )}
        </SelectContent>
      </Select>
      {selectedModel && (
        <div className="p-3 rounded-md border bg-muted/50 space-y-2">
          <div className="flex items-start gap-2">
            <Cpu className="w-4 h-4 mt-0.5 text-muted-foreground" />
            <div className="flex-1">
              <p className="text-sm font-medium">{selectedModel.name}</p>
              <p className="text-xs text-muted-foreground">{selectedModel.description}</p>
            </div>
          </div>
          <div className="flex items-center gap-2 text-xs text-muted-foreground">
            <Database className="w-3 h-3" />
            <span>
              Context: {selectedModel.contextWindow?.toLocaleString()} tokens
            </span>
            {selectedModel.maxTokens && (
              <span>| Max Output: {selectedModel.maxTokens.toLocaleString()} tokens</span>
            )}
          </div>
        </div>
      )}
    </div>
  );
};
