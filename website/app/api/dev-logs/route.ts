import { NextRequest, NextResponse } from 'next/server';

interface LogBatch {
  logs: Array<{
    message: string;
    stack?: string;
    url: string;
    timestamp: string;
  }>;
}

export async function POST(request: NextRequest) {
  // Only accept logs in development mode
  if (process.env.NODE_ENV !== 'development') {
    return NextResponse.json({ error: 'Not available in production' }, { status: 403 });
  }

  try {
    const body: LogBatch = await request.json();

    for (const log of body.logs) {
      // Format log for server console output
      const timestamp = new Date(log.timestamp).toLocaleTimeString('en-US', {
        hour12: false,
        hour: '2-digit',
        minute: '2-digit',
        second: '2-digit',
        fractionalSecondDigits: 3,
      });

      const url = new URL(log.url);
      const path = url.pathname + url.search;

      // ANSI color codes for terminal output
      const reset = '\u001b[0m';
      const red = '\u001b[31m';
      const yellow = '\u001b[33m';
      const dim = '\u001b[2m';
      const bright = '\u001b[1m';

      console.log(
        `${bright}${red}[BROWSER ERROR]${reset} ${dim}${timestamp}${reset} ${yellow}${path}${reset}`
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

    return new NextResponse(null, { status: 204 });
  } catch {
    return NextResponse.json({ error: 'Invalid request' }, { status: 400 });
  }
}
