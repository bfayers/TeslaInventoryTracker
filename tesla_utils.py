import json
from typing import List, Tuple
import requests

BASE_URL = "https://www.tesla.com/inventory/api/v4/inventory-results"


class Car:
    def __init__(self, data, loaded_cars={}):
        self.data = data

        # Extract the data
        # Basic car details
        self.model = data["Model"]
        self.year = data["Year"]

        # Car Registration Details
        self.vin = data["VIN"]
        self.registration = data["RegistrationDetails"]["LicensePlateNumber"]
        self.first_registration = data["FirstRegistrationDate"]

        # Car Details (Odomoter, Trim, Options)
        self.odometer = data["Odometer"]
        self.odometer_type = data["OdometerType"]
        self.trim = data["TrimName"]
        self.options = [
            option["name"] for option in data["OptionCodeSpecs"]["C_OPTS"]["options"]
        ]

        # Car Location (Tesla Centre)
        self.location = data["VrlName"]

        # Car Price & Link
        self.price = data["InventoryPrice"]
        self.link = f"https://www.tesla.com/en_GB/{self.model}/order/{self.vin}"

        # Check if the car has photos
        if len(data["VehiclePhotos"]) > 0:
            self.photos = [photo["imageUrl"] for photo in data["VehiclePhotos"]]
        else:
            self.photos = None

        # Value to be saved
        self.save_data = {
            "vin": self.vin,
            "price": self.price,
            "photos": True if self.photos else False,
        }

        # Compare to already saved data
        self.is_new = self.vin not in loaded_cars
        self.price_changed_since_last = False
        self.price_change = 0
        self.photos_added_since_last = False

        if self.vin in loaded_cars:
            # Check if the price has changed
            if loaded_cars[self.vin]["price"] != self.price:
                self.price_changed_since_last = True
                self.price_change = self.price - loaded_cars[self.vin]["price"]
            # Check if the photos have been added
            if not loaded_cars[self.vin]["photos"] and self.photos:
                self.photos_added_since_last = True

    def __str__(self):
        return (
            f"{self.year} {self.model} {self.trim} - £{self.price}\n"
            f"{self.options}\n{self.location}\n"
            f"New: {self.is_new}\n"
            f"Price changed since last: {self.price_changed_since_last}: {self.price_change}\n"
            f"Photos added since last: {self.photos_added_since_last}"
        )


def get_tesla_inventory(
    model: str = "m3", years: List = ["2024", "2025"], trims: List = ["LRRWD", "LRAWD"]
) -> List[Car]:
    # Load saved cars data
    with open("cars.json", "r") as f:
        loaded_cars = json.load(f)

    # Build query string
    query = {
        "query": {
            "model": model,
            "condition": "used",
            "options": {"Year": years, "TRIM": trims},
            "arrangeby": "Year",
            "order": "desc",
            "market": "GB",
        }
    }

    # Get the inventory data
    response = requests.get(
        BASE_URL,
        params={"query": json.dumps(query)},
        headers={
            "Accept": "application/json",
            "User-Agent": "Mozilla/5.0 (Macintosh; Intel Mac OS X 10.15; rv:135.0) Gecko/20100101 Firefox/135.0",
        },
    )
    data = response.json()

    inventory = [Car(vehicle, loaded_cars) for vehicle in data["results"]]

    with open("cars.json", "w") as f:
        output = {}
        for car in inventory:
            output[car.vin] = car.save_data
        json.dump(output, f, indent=4)

    return inventory


def identify_new_and_changed(inventory: List[Car]) -> Tuple[List[Car], List[Car]]:
    new_cars = []
    changed_cars = []

    for car in inventory:
        if car.is_new:
            new_cars.append(car)
        if car.price_changed_since_last or car.photos_added_since_last:
            changed_cars.append(car)

    return new_cars, changed_cars


if __name__ == "__main__":
    inventory = get_tesla_inventory()

    new_cars, changed_cars = identify_new_and_changed(inventory)

    print("New Cars:")
    for car in new_cars:
        print(car)
        print(car.link)
        print("=" * 55)

    print("Changed Cars:")
    for car in changed_cars:
        print(car)
        print(car.link)
        print("=" * 55)
