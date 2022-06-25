import sys
import unittest
from time import sleep
from selenium import webdriver
from selenium.webdriver.common.keys import Keys
from selenium.webdriver.common.by import By
from selenium.webdriver.support.ui import WebDriverWait
from selenium.webdriver.support import expected_conditions as EC

class TestKeenTune_UI_normal(unittest.TestCase):
    @classmethod
    def setUpClass(self,no_ui=False) -> None:
        if 'linux' in sys.platform:
            option = webdriver.ChromeOptions()
            option.add_argument('headless')
            option.add_argument('no-sandbox')
            option.add_argument('--start-maximized')
            option.add_argument('--disable-gpu')
            option.add_argument('--window-size=1920,1080')
            self.driver = webdriver.Chrome(options=option)
            self.driver.implicitly_wait(3)
            self.wait = WebDriverWait(self.driver,20,0.5)

        else:
            if no_ui:
                option = webdriver.ChromeOptions()
                option.add_argument('headless')
                option.add_argument('--start-maximized')
                self.driver = webdriver.Chrome(chrome_options=option)
            else:
                self.driver = webdriver.Chrome()
                self.driver.maximize_window()

        self.driver.get("http://39.102.55.119:8082/list/static-page")
        return self.driver

    @classmethod
    def tearDownClass(self) -> None:
        self.driver.get("http://39.102.55.119:8082/list/static-page")
        for i in range(9):
            first_text = self.driver.find_element(By.XPATH,'//tr[@data-row-key="1"]//td[1]//div[1]//span[1]').text
            if first_text != 'cpu_high_load.conf':
                self.wait.until(EC.element_to_be_clickable((By.XPATH,'//tr[@data-row-key="1"]//td[4]//div[1]//div[1]/span'))).click()
                self.wait.until(EC.element_to_be_clickable((By.XPATH,'//div[@class="ant-popover-buttons"]/button[2]/span'))).click()
            else:
                break
        self.driver.quit()


    def test_copyfile(self):
        self.wait.until(EC.element_to_be_clickable((By.XPATH,'//tr[@data-row-key="1"]/td[4]//div[2]'))).click()
        self.driver.find_element(By.ID, "name").send_keys(Keys.CONTROL, "a")
        self.driver.find_element(By.ID, "name").send_keys(Keys.BACKSPACE)
        self.driver.find_element(By.ID, "name").send_keys("1")
        self.driver.find_element(By.XPATH,'//div[@class="ant-modal-mask"]/../div[2]/div[1]/div[2]/div[3]/div[1]/div[2]').click()
        ele_copy = self.driver.find_element(By.XPATH, '//tr[@data-row-key="1"]//td[1]')
        sleep(1)
        assert "1.conf" in ele_copy.text

    def test_creatfile(self):
        self.wait.until(EC.element_to_be_clickable((By.XPATH,'//button[@class="ant-btn ant-btn-primary"]'))).click()
        self.driver.find_element(By.ID, "name").send_keys("11")
        self.driver.find_element(By.ID, "info").send_keys("[my.con]\ninnodb_file_per_table: 1")
        self.driver.find_element(By.XPATH,
                                 '//div[@class="ant-modal-mask"]/../div[2]/div[1]/div[2]/div[3]/div[1]/div[2]').click()
        sleep(1)
        ele_creat = self.driver.find_element(By.XPATH, '//tr[@data-row-key="2"]//td[1]')
        sleep(1)
        assert "11.conf" in ele_creat.text

    def test_checkfile(self):
        self.wait.until(EC.element_to_be_clickable((By.XPATH,'//tr[@data-row-key="2"]/td[1]//span'))).click()
        sleep(1)
        ele_checkfile = self.driver.find_element(By.XPATH,'//div[@class="CodeMirror-code"]')
        assert "[my.con]" in ele_checkfile.text
        self.driver.find_element(By.XPATH,
                                 '//div[@class="ant-modal-mask"]/../div[2]/div[1]/div[2]/div[3]/button').click()

    def test_editor(self):
        self.wait.until(EC.element_to_be_clickable((By.XPATH,'//tr[@data-row-key="2"]/td[4]//div[3]/span'))).click()        
        self.driver.find_element(By.ID, "name").send_keys(Keys.CONTROL, "a")
        self.driver.find_element(By.ID, "name").send_keys(Keys.BACKSPACE)
        self.driver.find_element(By.ID, "name").send_keys("111")
        self.driver.find_element(By.XPATH,
                                 '//div[@class="ant-modal-mask"]/../div[2]/div[1]/div[2]/div[3]/div[1]/div[2]').click()
        ele_editor = self.driver.find_element(By.XPATH, '//tr[@data-row-key="2"]//td[1]')
        sleep(1)
        assert "111.conf" in ele_editor.text

    def test_set_group(self):
        self.wait.until(EC.element_to_be_clickable((By.XPATH,'//tr[@data-row-key="1"]//td[4]//div[5]'))).click()
        self.driver.find_element(By.XPATH, '//label[@class="ant-radio-wrapper"]//span[2]').click()
        self.driver.find_element(By.XPATH, '//div[3]/div/div[2]/button/span').click()
        sleep(1)
        ele_set = self.driver.find_element(By.XPATH,'//tr[@data-row-key="1"]/td[3]')
        assert "[target-group-1]" in ele_set.text

    def test_restore(self):
        self.wait.until(EC.element_to_be_clickable((By.XPATH,'//tr[@data-row-key="1"]/td[4]//div[4]'))).click()
        sleep(1)
        ele_set = self.driver.find_element(By.XPATH, '//tr[@data-row-key="1"]/td[3]')
        assert "[target-group-1]\nTARGET_IP = localhost" not in ele_set.text

    def test_deletefile(self):
        for i in range(9):
            first_text = self.driver.find_element(By.XPATH,'//tr[@data-row-key="1"]//td[1]//div[1]//span[1]').text
            if first_text != 'cpu_high_load.conf':
                self.wait.until(EC.element_to_be_clickable((By.XPATH,'//tr[@data-row-key="1"]//td[4]//div[1]//div[1]/span'))).click()
                self.wait.until(EC.element_to_be_clickable((By.XPATH,'//div[@class="ant-popover-buttons"]/button[2]/span'))).click()
            else:
                break
        ele_copy = self.driver.find_element(By.XPATH, '//tr[@data-row-key="1"]//td[1]')
        assert "1.conf" not in ele_copy.text

    def test_language_switch(self):
        self.wait.until(EC.element_to_be_clickable((By.XPATH,
                                                '//div[@class="ant-space ant-space-horizontal ant-space-align-center right___3L8KG"]/div/div/img'))).click()
        ele_language1 = self.driver.find_element(By.XPATH,
                                                 '//div[@class="ant-pro-table-list-toolbar-title"]')
        sleep(1)
        assert "List of Expert Knowledge Based Tuning Profiles" in ele_language1.text

    def test_refresh(self):
        self.wait.until(EC.element_to_be_clickable((By.XPATH,'//div[@class="ant-space ant-space-horizontal ant-space-align-center '
                                                     'ant-pro-table-list-toolbar-setting-items"]//div[1]/div'))).click()

    def test_set_list(self):
        self.wait.until(EC.element_to_be_clickable((By.XPATH,
                                            '//div[@class="ant-space ant-space-horizontal ant-space-align-center ant-pro-table-list-toolbar-setting-items"]//div[2]'))).click()
        self.wait.until(EC.element_to_be_clickable((By.XPATH,'//div[@class="ant-tree-list-holder-inner"]/div[1]//span[4]'))).click()
        sleep(1)
        ele = self.driver.find_element(By.XPATH, '//thead[@class="ant-table-thead"]')
        assert "Profile Name" not in ele.text
        self.driver.find_element(By.XPATH, '//a[@class="ant-pro-table-column-setting-action-rest-button"]').click()
        assert "Profile Name" in ele.text
