import pandas as pd
import glob
import os
import requests
import config
from models import AddressQueryList
from pydantic import parse_obj_as
from datetime import datetime


input_file_path = os.path.join(os.path.dirname(__file__), "data", "input")
output_file_path = os.path.join(os.path.dirname(__file__), "data", "output", "output.csv")


def get_input_data():
    """This method is responsible for reading files from
    the directory and creating a single dataframe"""
    files = glob.glob(input_file_path + "/*.*")
    df = []
    if len(files) > 0:
        for f in files:
            print(f"Reading the file {f}")
            if f.endswith("csv") or f.endswith("txt"):
                dataframe = pd.read_csv(f, encoding='utf-8', sep="\\n", engine="python")
            elif f.endswith("xlsx"):
                dataframe = pd.read_excel(f, engine='openpyxl')
            elif f.endswith("xls"):
                dataframe = pd.read_excel(f)
            else:
                continue
            df.append(dataframe)
    else:
        raise Exception(
            "No input file present in input data directory "
            + input_file_path
        )
    return pd.concat(df)


def get_processed_input_data(df: None):
    """This method is responsible for reading dataframe having one column
    and transforming it to JSON structure"""
    records = df.to_dict("records")
    job_titles = parse_obj_as(AddressQueryList, records)
    data = job_titles.json()
    return data


def create_post_request(end_point_url, pay_load):
    print("Creating post request to " + end_point_url)
    try:
        request_output = requests.post(url=end_point_url, data=pay_load)
    except requests.Timeout:
        pass
    except requests.ConnectionError:
        pass
    response = None
    if request_output is not None:
        if request_output.status_code == 200:
            response = request_output.json()
    return response


def parse_response_output(response):
    return pd.DataFrame(response["Outputs"])


def write_output_data(df: None):
    """This method is responsible for write dataframe into CSV files"""
    print("writing output files in " + output_file_path)
    df.to_csv(output_file_path, index=False)


def process(data, env):
    if env == "dev":
        response_data = create_post_request(
            end_point_url=config.JOB_LEVEL_API_ENDPOINT_DEV, pay_load=data
        )
    else:
        response_data = create_post_request(
            end_point_url=config.ClientConfig.JOB_LEVEL_API_ENDPOINT_PROD, pay_load=data
        )
    if response_data is not None:
        address_parser_df = parse_response_output(response_data)
        return address_parser_df
    else:
        raise Exception("Exception while parsing the address")


def main():
    start_time = datetime.now()
    env_val = input("Please select environment, just press Enter key for prod environment or type dev for selecting "
                    "dev environment environment :: ")
    if env_val == "dev":
        print("dev environment")
    elif env_val == "prod":
        print("prod environment")
    else:
        print("invalid environment")
        print("exiting the execution")
        exit()

    df = get_input_data()
    print(df.shape)
    data = get_processed_input_data(df)
    address_df = process(data, env_val)
    write_output_data(address_df)
    end_time = datetime.now()
    print('Response time  is: {}'.format(end_time - start_time))


if __name__ == '__main__':
    main()
