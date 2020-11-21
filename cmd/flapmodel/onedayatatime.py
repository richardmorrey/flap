import argparse
import datetime
import os
parser = argparse.ArgumentParser()
parser.add_argument('startdate', type=lambda s: datetime.datetime.strptime(s, '%Y-%m-%d'))
parser.add_argument('daystorun', type=int)
args = parser.parse_args()
os.system("./flapmodel warm")
day = args.startdate
for i in range(1,args.daystorun): 
    
    # Run model for one day
    os.system("./flapmodel runoneday " + day.strftime("%Y-%m-%d"))

    # Move to next day
    day = day + datetime.timedelta(days=1)

# Report results
os.system("./flapmodel report")
