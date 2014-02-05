set :user, 'rob'
set :domain, 'wdcboard.com'
set :deploy_to, '/home/rob/wdcboard'

# Manually create these paths in shared/ (eg: shared/config/database.yml) in your server.
# They will be linked in the 'deploy:link_shared_paths' step.
set :shared_paths, []

# Optional settings:
#   set :user, 'foobar'    # Username in the server to SSH to.
#   set :port, '30000'     # SSH port number.

# This task is the environment that is loaded for most commands, such as
# `mina deploy` or `mina rake`.
task :environment do
end

# Put any custom mkdir's in here for when `mina setup` is ran.
# For Rails apps, we'll make some of the shared paths that are shared between
# all releases.
task :setup => :environment do
end

task :compile do
  puts "building sources"
  system "cd src/github.com/robmerrell/wdcboard && gxc build linux/amd64"
end

task :upload_binary do
  puts "uploading binary"
  system "scp src/github.com/robmerrell/wdcboard/wdcboard-linux-amd64 #{user}@#{domain}:#{deploy_to}/tmp/wdcboard"
  queue echo_cmd %[mv "#{deploy_to}/tmp/wdcboard" "wdcboard"]
end

task :upload_resources do
  puts "uploading resources"
  system "scp -r src/github.com/robmerrell/wdcboard/resources #{user}@#{domain}:#{deploy_to}/tmp/resources"
  queue echo_cmd %[mv "#{deploy_to}/tmp/resources" "resources"]
end

desc "Deploys the current version to the server."
task :deploy => :environment do
  deploy do
    invoke :compile
    invoke :upload_binary
    invoke :upload_resources

    to :launch do
      queue "sudo restart wdcboard"
    end
  end
end

