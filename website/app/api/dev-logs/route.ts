import { NextRequest } from 'next/server';

interface LogEntry {
  message: string;
  stack?: string;
  url: string;
  timestamp: string;
}

interface LogBatch {
  logs: LogEntry[];
}

export async function POST(request: NextRequest) {
  // Only accept logs in development mode
  if (process.env.NODE_ENV !== 'development') {
    return Response.json({ error: 'Not available in production' }, { status: 403 });
  }

  try {
    const body: LogBatch = await request.json();

    for (const log of body.logs) {
      // Format log for server console output
      const date = new Date(log.timestamp);
      const timestamp = date.toLocaleTimeString('en-US', {
        hour12: false,
        hour: '2-digit',
        minute: '2-digit',
        second: '2-digit',
      });
      const ms = String(date.getMilliseconds()).padStart(3, '0');

      const url = new URL(log.url);
      const path = url.pathname + url.search;

      // ANSI color codes for terminal output
      const reset = '\x1b[0m';
      const red = '\x1b[31m';
      const yellow = '\x1b[33m';
      const dim = '\x1b[2m';
      const bright = '\x1b[1m';

      console.log(
        `${bright}${red}[BROWSER ERROR]${reset} ${dim}${timestamp}.${ms}${reset} ${yellow}${path}${reset}`
      );
      console.log(`  ${dim}Message:${reset} ${log.message}`);

      if (log.stack) {
        // Format stack trace with indentation
        const lines = log.stack.split('\n');
        console.log(`  ${dim}Stack:${reset}`);
        for (const line of lines) {
          console.log(`    ${dim}${line}${reset}`);
        }
      }
    }

    return new Response(null, { status: 204 });
  } catch {
    return Response.json({ error: 'Invalid request' }, { status: 400 });
  }
}
