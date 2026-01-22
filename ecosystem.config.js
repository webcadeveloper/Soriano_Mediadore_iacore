module.exports = {
  apps: [
    // === PRODUCCION ===
    {
      name: 'soriano-api',
      cwd: '/opt/soriano/backend',
      script: './soriano-api',
      instances: 1,
      autorestart: true,
      watch: false,
      max_memory_restart: '500M',
      env: {
        PORT: 8080,
        GIN_MODE: 'release'
      },
      error_file: '/opt/soriano/logs/api-error.log',
      out_file: '/opt/soriano/logs/api-out.log',
      log_date_format: 'YYYY-MM-DD HH:mm:ss Z'
    },
    {
      name: 'soriano-frontend',
      cwd: '/opt/soriano/frontend',
      script: '/usr/bin/serve',
      args: '-s . -l 4200',
      instances: 1,
      autorestart: true,
      watch: false,
      max_memory_restart: '200M',
      error_file: '/opt/soriano/logs/frontend-error.log',
      out_file: '/opt/soriano/logs/frontend-out.log',
      log_date_format: 'YYYY-MM-DD HH:mm:ss Z'
    },
    // === TEST ===
    {
      name: 'soriano-test-api',
      cwd: '/opt/soriano-test/backend',
      script: './soriano-api',
      instances: 1,
      autorestart: true,
      watch: false,
      max_memory_restart: '500M',
      env: {
        PORT: 8081,
        GIN_MODE: 'release'
      },
      error_file: '/opt/soriano-test/logs/api-error.log',
      out_file: '/opt/soriano-test/logs/api-out.log',
      log_date_format: 'YYYY-MM-DD HH:mm:ss Z'
    },
    {
      name: 'soriano-test-frontend',
      cwd: '/opt/soriano-test/frontend',
      script: '/usr/bin/serve',
      args: '-s . -l 4201',
      instances: 1,
      autorestart: true,
      watch: false,
      max_memory_restart: '200M',
      error_file: '/opt/soriano-test/logs/frontend-error.log',
      out_file: '/opt/soriano-test/logs/frontend-out.log',
      log_date_format: 'YYYY-MM-DD HH:mm:ss Z'
    }
  ]
};
