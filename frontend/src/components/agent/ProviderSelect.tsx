import React from "react";
import { ProviderSelectProps } from "./types";
import { PROVIDERS } from "./constants";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { Label } from "@/components/ui/label";
import { Badge } from "@/components/ui/badge";
import { Cpu, Lock, Unlock, Eye, EyeOff, Zap } from "lucide-react";

export const ProviderSelect: React.FC<ProviderSelectProps> = ({
  value,
  onChange,
  disabled = false,
}) => {
  const selectedProvider = PROVIDERS.find((p) => p.id === value);

  const getCapabilityIcon = (capability: string, enabled: boolean) => {
    if (capability === "toolUse") {
      return enabled ? <Unlock className="w-3 h-3" /> : <Lock className="w-3 h-3" />;
    }
    if (capability === "vision") {
      return enabled ? <Eye className="w-3 h-3" /> : <EyeOff className="w-3 h-3" />;
    }
    if (capability === "streaming") {
      return <Zap className="w-3 h-3" />;
    }
    return null;
  };

  return (
    <div className="space-y-2">
      <Label htmlFor="provider">Provider</Label>
      <Select
        value={value}
        onValueChange={onChange}
        disabled={disabled}
      >
        <SelectTrigger id="provider">
          <SelectValue placeholder="Select a provider" />
        </SelectTrigger>
        <SelectContent>
          {PROVIDERS.map((provider) => (
            <SelectItem key={provider.id} value={provider.id}>
              <div className="flex items-center justify-between w-full">
                <span>{provider.name}</span>
                <div className="flex items-center gap-1 ml-2">
                  {getCapabilityIcon("toolUse", provider.capabilities.toolUse)}
                  {getCapabilityIcon("vision", provider.capabilities.vision)}
                  {getCapabilityIcon("streaming", provider.capabilities.streaming)}
                </div>
              </div>
            </SelectItem>
          ))}
        </SelectContent>
      </Select>
      {selectedProvider && (
        <div className="flex flex-wrap gap-2 text-xs">
          <Badge variant="outline">
            <Cpu className="w-3 h-3 mr-1" />
            {selectedProvider.capabilities.toolUse ? "Tool Use" : "No Tools"}
          </Badge>
          {selectedProvider.capabilities.vision && (
            <Badge variant="outline">
              <Eye className="w-3 h-3 mr-1" />
              Vision
            </Badge>
          )}
          <Badge variant="outline">
            <Zap className="w-3 h-3 mr-1" />
            {selectedProvider.capabilities.streaming ? "Streaming" : "No Stream"}
          </Badge>
        </div>
      )}
      {selectedProvider && selectedProvider.baseURL && (
        <p className="text-xs text-muted-foreground">
          Base URL: {selectedProvider.baseURL}
        </p>
      )}
    </div>
  );
};
