from tesla_utils import get_tesla_inventory, identify_new_and_changed
import requests
import os

webhook_url = os.getenv("DISCORD_WEBHOOK_URL")
new_car_thread = os.getenv("DISCORD_NEW_CAR_THREAD")
changed_car_thread = os.getenv("DISCORD_CHANGED_CAR_THREAD")


def send_car_to_discord(car, thread_id):
    data = {
        "embeds": [
            {
                "title": f"{car.year} {car.model} {car.trim} - £{car.price:,}",
                "url": car.link,
                "color": 5814783,
                "fields": [
                    {"name": "VIN", "value": car.vin, "inline": True},
                    {"name": "Plate", "value": car.registration, "inline": True},
                    {
                        "name": "Options",
                        "value": "\n".join(car.options),
                    },
                    {
                        "name": "Odometer",
                        "value": f"{car.odometer} {car.odometer_type}",
                        "inline": True,
                    },
                ],
                "author": {"name": car.location},
                "image": {},
                "thumbnail": {},
            },
        ],
    }
    # If the car has photos, add them to the embed
    if car.photos:
        data["embeds"][0]["thumbnail"]["url"] = car.photos[0]
        data["embeds"][0]["image"]["url"] = car.photos[1]
        for photo in car.photos[2:5]:
            data["embeds"].append(
                {
                    "url": car.link,
                    "image": {"url": photo},
                }
            )
    # If this is an update to an existing car, add a field explaining if the update is a price change, photos added, or both.
    if not car.is_new:
        if car.price_changed_since_last:
            data["embeds"][0]["fields"].append(
                {
                    "name": "Price Change",
                    "value": f"{'+' if car.price_change > 0 else ''}{car.price_change:,}",
                    "inline": True,
                }
            )
        if car.photos_added_since_last:
            data["embeds"][0]["fields"].append(
                {
                    "name": "Photos Added",
                    "value": "Yes",
                    "inline": True,
                }
            )
    requests.post(webhook_url, json=data, params={"thread_id": thread_id})
    print(f"Sent {car.vin} to Discord")


# Get the inventory
inventory = get_tesla_inventory(trims=["LRRWD", "LRAWD"])
# Identify new and changed cars
new_cars, changed_cars = identify_new_and_changed(inventory)

# Send the new cars to Discord
for car in new_cars:
    send_car_to_discord(car, thread_id=new_car_thread)

for car in changed_cars:
    send_car_to_discord(car, thread_id=changed_car_thread)
