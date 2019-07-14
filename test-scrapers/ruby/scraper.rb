require 'scraperwiki'
require "open-uri"

puts "Running scraper"
# Check that we can use the scraperwiki library
ScraperWiki.save_sqlite(["name"], {"name" => "susan", "occupation" => "software developer", "time" => Time.now})
# Check that mitmproxy gives a working certificate for SSL connections
# If the certificate isn't valid it should throw an exception
# open("https://morph.io") do |f|
#   raise "Unexpected result" unless f.read.include?("Hassle-free web scraping.")
# end
# Check that output streaming works (in which case you should see each line one second apart
# rather than getting all the lines at the end)
puts "Sleeping for 5 seconds"
(1..5).each do |i|
  puts "#{i}..."
  sleep 1
end
puts "Finished"
