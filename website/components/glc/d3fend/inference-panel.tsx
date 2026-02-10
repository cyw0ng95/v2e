/**
 * D3FEND Inference Panel
 *
 * Shows sensor coverage, mitigation suggestions, and weakness analysis
 * for the entire graph.
 */

import React, { useMemo } from 'react';
import * as LucideIcons from 'lucide-react';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Badge } from '@/components/ui/badge';
import { Button } from '@/components/ui/button';
import { Separator } from '@/components/ui/separator';
import {
  getSensorCoverage,
  getGraphInferences,
  type SensorDetection,
  type InferenceResult,
} from '@/lib/glc/d3fend';
import type { Node, Edge } from '@xyflow/react';

// ============================================================================
// Props
// ============================================================================

interface D3FENDInferencePanelProps {
  nodes: Node[];
  edges: Edge[];
  isOpen: boolean;
  onClose: () => void;
}

// ============================================================================
// Helper Components
// ============================================================================

const CoverageGauge: React.FC<{ score: number }> = ({ score }) => {
  const getColor = (score: number) => {
    if (score >= 80) return 'bg-green-500';
    if (score >= 60) return 'bg-blue-500';
    if (score >= 40) return 'bg-yellow-500';
    return 'bg-red-500';
  };

  return (
    <div className="flex items-center gap-3">
      <div className="relative w-16 h-16">
        <svg className="w-full h-full transform -rotate-90">
          <circle
            cx="32"
            cy="32"
            r="28"
            stroke="currentColor"
            strokeWidth="6"
            fill="none"
            className="text-muted"
          />
          <circle
            cx="32"
            cy="32"
            r="28"
            stroke="currentColor"
            strokeWidth="6"
            fill="none"
            strokeDasharray={`${2 * Math.PI * 28}`}
            strokeDashoffset={`${2 * Math.PI * 28 * (1 - score / 100)}`}
            className={getColor(score)}
          />
        </svg>
        <div className="absolute inset-0 flex items-center justify-center">
          <span className="text-lg font-bold">{score}</span>
        </div>
      </div>
      <div>
        <div className="text-sm font-medium">Sensor Coverage</div>
        <div className="text-xs text-muted-foreground">
          {score >= 80 ? 'Excellent' : score >= 60 ? 'Good' : score >= 40 ? 'Fair' : 'Poor'}
        </div>
      </div>
    </div>
  );
};

const SensorCard: React.FC<{ detection: SensorDetection }> = ({ detection }) => {
  return (
    <div className="flex items-start gap-3 p-3 rounded-lg border bg-card">
      <LucideIcons.Radar className="h-5 w-5 text-primary mt-0.5" />
      <div className="flex-1 min-w-0">
        <div className="font-medium text-sm">{detection.nodeType}</div>
        <div className="text-xs text-muted-foreground mt-1">
          {detection.sensors.slice(0, 3).join(', ')}
          {detection.sensors.length > 3 && ` +${detection.sensors.length - 3} more`}
        </div>
        <Badge variant="outline" className="mt-2">
          {detection.coverageScore}% coverage
        </Badge>
      </div>
    </div>
  );
};

const InferenceCard: React.FC<{ inference: InferenceResult }> = ({ inference }) => {
  const getIcon = (type: InferenceResult['type']) => {
    const icons: Record<InferenceResult['type'], React.ReactNode> = {
      sensor: <LucideIcons.Radar className="h-4 w-4" />,
      mitigation: <LucideIcons.Shield className="h-4 w-4" />,
      detection: <LucideIcons.Eye className="h-4 w-4" />,
      weakness: <LucideIcons.Bug className="h-4 w-4" />,
    };
    return icons[type];
  };

  const getSeverityColor = (severity: InferenceResult['severity']) => {
    const colors: Record<InferenceResult['severity'], string> = {
      critical: 'bg-red-500',
      high: 'bg-orange-500',
      medium: 'bg-yellow-500',
      low: 'bg-blue-500',
      info: 'bg-gray-500',
    };
    return colors[severity];
  };

  return (
    <div className="flex items-start gap-3 p-3 rounded-lg border hover:bg-accent/50 transition-colors cursor-pointer">
      <div className={`p-2 rounded-full ${getSeverityColor(inference.severity)} text-white`}>
        {getIcon(inference.type)}
      </div>
      <div className="flex-1 min-w-0">
        <div className="font-medium text-sm">{inference.title}</div>
        <p className="text-xs text-muted-foreground line-clamp-2 mt-1">
          {inference.description}
        </p>
        {inference.confidence > 0 && (
          <div className="flex items-center gap-1 mt-2">
            <div className="h-1.5 w-12 bg-muted rounded-full overflow-hidden">
              <div
                className="h-full bg-primary transition-all"
                style={{ width: `${inference.confidence}%` }}
              />
            </div>
            <span className="text-xs text-muted-foreground">
              {inference.confidence}% confidence
            </span>
          </div>
        )}
      </div>
    </div>
  );
};

// ============================================================================
// D3FEND Inference Panel Component
// ============================================================================

export const D3FENDInferencePanel: React.FC<D3FENDInferencePanelProps> = ({
  nodes,
  edges,
  isOpen,
  onClose,
}) => {
  if (!isOpen) return null;

  const { score, detections } = useMemo(() => getSensorCoverage(nodes, edges), [nodes, edges]);
  const inferences = useMemo(() => getGraphInferences(nodes, edges), [nodes, edges]);

  const criticalInferences = inferences.filter(i => i.severity === 'critical');
  const highInferences = inferences.filter(i => i.severity === 'high');

  return (
    <Card className="w-96 max-h-[80vh] overflow-y-auto shadow-lg">
      <CardHeader>
        <div className="flex items-center justify-between">
          <CardTitle className="text-lg flex items-center gap-2">
            <LucideIcons.Activity className="h-5 w-5" />
            D3FEND Analysis
          </CardTitle>
          <Button variant="ghost" size="icon" onClick={onClose}>
            <LucideIcons.X className="h-4 w-4" />
          </Button>
        </div>
        <CardDescription>
          Sensor coverage and inference results
        </CardDescription>
      </CardHeader>

      <CardContent className="space-y-4">
        {/* Sensor Coverage */}
        <div>
          <h3 className="text-sm font-medium mb-3">Sensor Coverage</h3>
          <CoverageGauge score={score} />
        </div>

        <Separator />

        {/* Active Sensors */}
        {detections.length > 0 && (
          <div>
            <h3 className="text-sm font-medium mb-3 flex items-center gap-2">
              <LucideIcons.Radar className="h-4 w-4" />
              Active Sensors ({detections.length})
            </h3>
            <div className="space-y-2">
              {detections.map(detection => (
                <SensorCard key={detection.nodeId} detection={detection} />
              ))}
            </div>
          </div>
        )}

        {detections.length > 0 && <Separator />}

        {/* Critical Issues */}
        {criticalInferences.length > 0 && (
          <div>
            <h3 className="text-sm font-medium mb-3 flex items-center gap-2">
              <LucideIcons.AlertCircle className="h-4 w-4 text-red-500" />
              Critical Issues ({criticalInferences.length})
            </h3>
            <div className="space-y-2">
              {criticalInferences.map(inference => (
                <InferenceCard key={inference.id} inference={inference} />
              ))}
            </div>
          </div>
        )}

        {highInferences.length > 0 && <Separator />}

        {/* High Priority */}
        {highInferences.length > 0 && (
          <div>
            <h3 className="text-sm font-medium mb-3 flex items-center gap-2">
              <LucideIcons.AlertTriangle className="h-4 w-4 text-orange-500" />
              High Priority ({highInferences.length})
            </h3>
            <div className="space-y-2">
              {highInferences.map(inference => (
                <InferenceCard key={inference.id} inference={inference} />
              ))}
            </div>
          </div>
        )}

        {/* No Issues */}
        {criticalInferences.length === 0 && highInferences.length === 0 && (
          <div className="text-center py-8">
            <LucideIcons.CheckCircle className="h-12 w-12 text-green-500 mx-auto mb-3" />
            <div className="font-medium">No Critical Issues</div>
            <p className="text-sm text-muted-foreground">
              Your graph looks good!
            </p>
          </div>
        )}
      </CardContent>

      <CardHeader className="pt-0">
        <Button variant="outline" className="w-full" onClick={onClose}>
          Close
        </Button>
      </CardHeader>
    </Card>
  );
};
