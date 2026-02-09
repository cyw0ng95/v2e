export class GLCError extends Error {
  public readonly code: string;
  public readonly context?: Record<string, any>;

  constructor(message: string, code: string = 'GLC_ERROR', context?: Record<string, any>) {
    super(message);
    this.name = 'GLCError';
    this.code = code;
    this.context = context;
  }
}

export class PresetValidationError extends GLCError {
  public readonly validationErrors: any[];

  constructor(message: string, validationErrors: any[] = [], context?: Record<string, any>) {
    super(message, 'PRESET_VALIDATION_ERROR', context);
    this.name = 'PresetValidationError';
    this.validationErrors = validationErrors;
  }
}

export class GraphValidationError extends GLCError {
  public readonly validationErrors: any[];

  constructor(message: string, validationErrors: any[] = [], context?: Record<string, any>) {
    super(message, 'GRAPH_VALIDATION_ERROR', context);
    this.name = 'GraphValidationError';
    this.validationErrors = validationErrors;
  }
}

export class StateError extends GLCError {
  constructor(message: string, context?: Record<string, any>) {
    super(message, 'STATE_ERROR', context);
    this.name = 'StateError';
  }
}

export class RPCTimeoutError extends GLCError {
  public readonly timeout: number;

  constructor(message: string, timeout: number, context?: Record<string, any>) {
    super(message, 'RPC_TIMEOUT_ERROR', context);
    this.name = 'RPCTimeoutError';
    this.timeout = timeout;
  }
}

export class NetworkError extends GLCError {
  constructor(message: string, context?: Record<string, any>) {
    super(message, 'NETWORK_ERROR', context);
    this.name = 'NetworkError';
  }
}

export class FileSystemError extends GLCError {
  constructor(message: string, context?: Record<string, any>) {
    super(message, 'FILE_SYSTEM_ERROR', context);
    this.name = 'FileSystemError';
  }
}

export class SerializationError extends GLCError {
  constructor(message: string, context?: Record<string, any>) {
    super(message, 'SERIALIZATION_ERROR', context);
    this.name = 'SerializationError';
  }
}

export const isGLCError = (error: unknown): error is GLCError => {
  return error instanceof GLCError;
};

export const getErrorCode = (error: unknown): string => {
  if (isGLCError(error)) {
    return error.code;
  }
  if (error instanceof Error) {
    return error.name;
  }
  return 'UNKNOWN_ERROR';
};

export const getErrorMessage = (error: unknown): string => {
  if (error instanceof Error) {
    return error.message;
  }
  return String(error);
};
