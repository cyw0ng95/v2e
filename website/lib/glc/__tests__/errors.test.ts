import { describe, it, expect } from '@jest/globals';
import {
  GLCError,
  PresetValidationError,
  GraphValidationError,
  StateError,
  RPCTimeoutError,
  NetworkError,
  FileSystemError,
  SerializationError,
  isGLCError,
  getErrorCode,
  getErrorMessage,
} from '../errors/error-types';
import { errorHandler, showError, showWarning, showInfo, clearErrorLogs, getErrorLogs } from '../errors/error-handler';

describe('Error Types', () => {
  it('should create GLCError with code', () => {
    const error = new GLCError('Test error', 'TEST_CODE');
    
    expect(error).toBeInstanceOf(GLCError);
    expect(error.message).toBe('Test error');
    expect(error.code).toBe('TEST_CODE');
    expect(error.name).toBe('GLCError');
  });

  it('should create PresetValidationError', () => {
    const validationErrors = [{ path: 'test', message: 'Invalid', code: 'INVALID' }];
    const error = new PresetValidationError('Validation failed', validationErrors);
    
    expect(error).toBeInstanceOf(PresetValidationError);
    expect(error.code).toBe('PRESET_VALIDATION_ERROR');
    expect(error.validationErrors).toEqual(validationErrors);
  });

  it('should create GraphValidationError', () => {
    const validationErrors = [{ path: 'test', message: 'Invalid', code: 'INVALID' }];
    const error = new GraphValidationError('Graph invalid', validationErrors);
    
    expect(error).toBeInstanceOf(GraphValidationError);
    expect(error.code).toBe('GRAPH_VALIDATION_ERROR');
    expect(error.validationErrors).toEqual(validationErrors);
  });

  it('should create StateError', () => {
    const error = new StateError('State error');
    
    expect(error).toBeInstanceOf(StateError);
    expect(error.code).toBe('STATE_ERROR');
  });

  it('should create RPCTimeoutError', () => {
    const error = new RPCTimeoutError('Timeout', 5000);
    
    expect(error).toBeInstanceOf(RPCTimeoutError);
    expect(error.code).toBe('RPC_TIMEOUT_ERROR');
    expect(error.timeout).toBe(5000);
  });

  it('should create NetworkError', () => {
    const error = new NetworkError('Network error');
    
    expect(error).toBeInstanceOf(NetworkError);
    expect(error.code).toBe('NETWORK_ERROR');
  });

  it('should create FileSystemError', () => {
    const error = new FileSystemError('File error');
    
    expect(error).toBeInstanceOf(FileSystemError);
    expect(error.code).toBe('FILE_SYSTEM_ERROR');
  });

  it('should create SerializationError', () => {
    const error = new SerializationError('Serialization error');
    
    expect(error).toBeInstanceOf(SerializationError);
    expect(error.code).toBe('SERIALIZATION_ERROR');
  });
});

describe('Error Utilities', () => {
  it('should identify GLCError', () => {
    const glcError = new GLCError('Test', 'TEST');
    const standardError = new Error('Test');
    
    expect(isGLCError(glcError)).toBe(true);
    expect(isGLCError(standardError)).toBe(false);
  });

  it('should get error code', () => {
    const glcError = new GLCError('Test', 'TEST_CODE');
    const standardError = new Error('Test');
    
    expect(getErrorCode(glcError)).toBe('TEST_CODE');
    expect(getErrorCode(standardError)).toBe('Error');
    expect(getErrorCode('string')).toBe('UNKNOWN_ERROR');
  });

  it('should get error message', () => {
    const error = new Error('Test message');
    const string = 'String error';
    
    expect(getErrorMessage(error)).toBe('Test message');
    expect(getErrorMessage(string)).toBe('String error');
  });
});

describe('Error Handler', () => {
  beforeEach(() => {
    clearErrorLogs();
  });

  it('should handle GLCError', () => {
    const error = new GLCError('Test error', 'TEST_ERROR');
    errorHandler.handleError(error);
    
    const logs = getErrorLogs();
    expect(logs).toHaveLength(1);
    expect(logs[0].code).toBe('TEST_ERROR');
    expect(logs[0].message).toBe('Test error');
  });

  it('should handle standard Error', () => {
    const error = new Error('Standard error');
    errorHandler.handleError(error);
    
    const logs = getErrorLogs();
    expect(logs).toHaveLength(1);
    expect(logs[0].code).toBe('Error');
    expect(logs[0].message).toBe('Standard error');
  });

  it('should handle unknown error', () => {
    errorHandler.handleError('Unknown error');
    
    const logs = getErrorLogs();
    expect(logs).toHaveLength(1);
    expect(logs[0].code).toBe('UNKNOWN_ERROR');
  });

  it('should include context in log', () => {
    const error = new GLCError('Test', 'TEST');
    const context = { action: 'test-action' };
    errorHandler.handleError(error, context);
    
    const logs = getErrorLogs();
    expect(logs[0].context).toEqual(context);
  });

  it('should store timestamp', () => {
    const before = new Date();
    const error = new GLCError('Test', 'TEST');
    errorHandler.handleError(error);
    const after = new Date();
    
    const logs = getErrorLogs();
    const timestamp = new Date(logs[0].timestamp);
    
    expect(timestamp).toBeInstanceOf(Date);
    expect(timestamp.getTime()).toBeGreaterThanOrEqual(before.getTime());
    expect(timestamp.getTime()).toBeLessThanOrEqual(after.getTime());
  });

  it('should limit log size', () => {
    for (let i = 0; i < 150; i++) {
      errorHandler.handleError(new Error(`Error ${i}`), 'TEST');
    }
    
    const logs = getErrorLogs();
    expect(logs.length).toBeLessThanOrEqual(100);
  });
});

describe('Error Handler - Convenience Methods', () => {
  beforeEach(() => {
    clearErrorLogs();
  });

  it('should log error', () => {
    showError('Test error');
    
    const logs = getErrorLogs();
    expect(logs).toHaveLength(1);
    expect(logs[0].message).toBe('Test error');
  });

  it('should show warning', () => {
    showWarning('Test warning');
    
    const logs = getErrorLogs();
    expect(logs).toHaveLength(1);
    expect(logs[0].message).toBe('Test warning');
  });

  it('should show info', () => {
    showInfo('Test info');
    
    const logs = getErrorLogs();
    expect(logs).toHaveLength(1);
    expect(logs[0].message).toBe('Test info');
  });

  it('should include context in convenience methods', () => {
    const context = { action: 'test' };
    showError('Test error', context);
    
    const logs = getErrorLogs();
    expect(logs[0].context).toEqual(context);
  });
});
