import os
import sys
import unittest
from time import sleep
from selenium import webdriver
from selenium.common.exceptions import NoSuchElementException
from selenium.webdriver.common.keys import Keys
from selenium.webdriver.common.by import By
from selenium.webdriver.support.ui import WebDriverWait
from selenium.webdriver.support import expected_conditions as EC

sys.path.append(os.path.abspath(os.path.join(os.getcwd(), "..")))

from common import keentuneInit


class TestKeenTuneUiSensitiveAbnormal(unittest.TestCase):
    @classmethod
    def setUpClass(self, no_ui=False) -> None:
        if 'linux' in sys.platform:
            option = webdriver.ChromeOptions()
            option.add_argument('headless')
            option.add_argument('no-sandbox')
            option.add_argument('--start-maximized')
            option.add_argument('--disable-gpu')
            option.add_argument('lang=zh_CN.UTF-8')
            option.add_argument('--window-size=1920,1080')
            self.driver = webdriver.Chrome(options=option)
            self.driver.implicitly_wait(3)
            self.wait = WebDriverWait(self.driver, 30, 0.5)

        else:
            if no_ui:
                option = webdriver.ChromeOptions()
                option.add_argument('headless')
                option.add_argument('--start-maximized')
                self.driver = webdriver.Chrome(chrome_options=option)
                self.wait = WebDriverWait(self.driver, 30, 0.5)
            else:
                self.driver = webdriver.Chrome()
                self.driver.maximize_window()
                self.wait = WebDriverWait(self.driver, 30, 0.5)
        #在智能参数调优页面创建任务
        keentuneInit(self)
        self.driver.get("http://{}:8082/list/tuning-task".format(self.web_ip))
        self.wait.until(EC.element_to_be_clickable((By.XPATH, '//button[@class="ant-btn ant-btn-default"]'))).click()
        self.wait.until(EC.visibility_of_element_located((By.ID, "name"))).send_keys(Keys.CONTROL, "a")
        self.wait.until(EC.visibility_of_element_located((By.ID, "name"))).send_keys(Keys.BACKSPACE)
        self.wait.until(EC.visibility_of_element_located((By.ID, "name"))).send_keys("auto_test_TPE")
        self.wait.until(EC.visibility_of_element_located((By.ID, "algorithm"))).send_keys(Keys.CONTROL, "a")
        self.wait.until(EC.visibility_of_element_located((By.ID, "algorithm"))).send_keys(Keys.BACKSPACE)
        self.wait.until(EC.visibility_of_element_located((By.ID, "algorithm"))).send_keys("TPE")
        self.wait.until(EC.visibility_of_element_located((By.ID, "algorithm"))).send_keys(Keys.ENTER)
        self.wait.until(EC.visibility_of_element_located((By.ID, "iteration"))).send_keys(Keys.CONTROL, "a")
        self.wait.until(EC.visibility_of_element_located((By.ID, "iteration"))).send_keys(Keys.BACKSPACE)
        self.wait.until(EC.visibility_of_element_located((By.ID, "iteration"))).send_keys(10)
        self.wait.until(EC.element_to_be_clickable(
            (By.XPATH, '//div[@class="ant-modal-mask"]/../div[2]/div[1]/div[2]/div[3]/div[1]/div[2]'))).click()
        # 等待任务执行完成，任务完成重新创建下一个，超时则结束
        for j in range(1, 9):
            sleep(35)
            self.driver.refresh()
            Total_Time = self.wait.until(
                EC.visibility_of_element_located((By.XPATH, '//tbody[@class="ant-table-tbody"]/tr[1]/td[11]'))).text
            if Total_Time != "-":
                break
            elif j == 8:
                self.assertNotIn("-", Total_Time)
        # 跳转到敏感参数页面
        self.driver.find_element(By.XPATH, '//div[@class="ant-pro-top-nav-header-logo"]//img').click()
        sleep(5)
        self.driver.find_element(By.XPATH,'//div[@class="list___y_nmN"]/div[3]//img').click()
        #在敏感参数页面基于智能参数页面创建的任务创建新任务
        # self.driver.get("http://{}:8082/list/sensitive-parameter".format(self.web_ip))
        value = self.driver.find_element(By.XPATH, '//div[@class="ant-pro-table-list-toolbar-title"]').text
        if "敏感参数识别任务记录" not in value:
            self.driver.find_element(By.XPATH,
                                 '//div[@class="ant-space ant-space-horizontal ant-space-align-center right___3L8KG"]/div/div/img').click()
        self.wait.until(EC.element_to_be_clickable((By.XPATH, '//button[@class="ant-btn ant-btn-default"]'))).click()
        self.wait.until(EC.visibility_of_element_located((By.ID, "name"))).send_keys(Keys.CONTROL, "a")
        self.wait.until(EC.visibility_of_element_located((By.ID, "name"))).send_keys(Keys.BACKSPACE)
        self.wait.until(EC.visibility_of_element_located((By.ID, "name"))).send_keys("auto_test_TPE")
        self.wait.until(EC.visibility_of_element_located((By.ID, "data"))).send_keys(Keys.CONTROL, "a")
        self.wait.until(EC.visibility_of_element_located((By.ID, "data"))).send_keys(Keys.BACKSPACE)
        self.wait.until(EC.visibility_of_element_located((By.ID, "data"))).send_keys("auto_test_TPE")
        self.wait.until(EC.visibility_of_element_located((By.ID, "data"))).send_keys(Keys.ENTER)
        self.wait.until(EC.visibility_of_element_located((By.ID, "algorithm"))).send_keys(Keys.CONTROL, "a")
        self.wait.until(EC.visibility_of_element_located((By.ID, "algorithm"))).send_keys(Keys.BACKSPACE)
        self.wait.until(EC.visibility_of_element_located((By.ID, "algorithm"))).send_keys("lasso")
        self.wait.until(EC.visibility_of_element_located((By.ID, "algorithm"))).send_keys(Keys.ENTER)
        self.wait.until(EC.visibility_of_element_located((By.ID, "trial"))).send_keys(Keys.CONTROL, "a")
        self.wait.until(EC.visibility_of_element_located((By.ID, "trial"))).send_keys(Keys.BACKSPACE)
        self.wait.until(EC.visibility_of_element_located((By.ID, "trial"))).send_keys("1")
        self.wait.until(EC.element_to_be_clickable(
            (By.XPATH, '//div[@class="ant-modal-mask"]/../div[2]/div[1]/div[2]/div[3]/div[1]/div[2]'))).click()

        #等待任务执行完成，任务完成重新创建下一个，超时则结束
        for j in range(1,9):
            sleep(35)
            self.driver.refresh()
            Total_Time = self.wait.until(EC.visibility_of_element_located((By.XPATH, '//tbody[@class="ant-table-tbody"]/tr[1]/td[9]'))).text
            if Total_Time != "-":
                break
            elif j == 8 :
                self.assertNotIn("-",Total_Time)

    @classmethod
    def tearDownClass(self) -> None:
        self.driver.get("http://{}:8082/list/tuning-task".format(self.web_ip))
        del_list = "auto_test_TPE"
        for i in range(9):
            first_text = self.wait.until(
                EC.visibility_of_element_located((By.XPATH, '//tbody[@class="ant-table-tbody"]/tr[1]/td[2]'))).text
            if first_text == del_list:
                self.wait.until(EC.element_to_be_clickable(
                    (By.XPATH, '//tbody[@class="ant-table-tbody"]/tr[1]/td[12]/div'))).click()
                self.wait.until(EC.element_to_be_clickable((By.XPATH,
                                                            '//ul[@class="ant-dropdown-menu ant-dropdown-menu-root ant-dropdown-menu-vertical ant-dropdown-menu-light"]/li[6]/span[1]'))).click()
                self.wait.until(EC.element_to_be_clickable((By.XPATH,
                                                            '//div[@class="ant-modal-confirm-body-wrapper"]//button[@class="ant-btn ant-btn-primary"]'))).click()
                sleep(1)
            else:
                break
        self.driver.find_element(By.XPATH, '//div[@class="ant-pro-top-nav-header-logo"]//img').click()
        sleep(5)
        self.driver.find_element(By.XPATH, '//div[@class="list___y_nmN"]/div[3]//img').click()
        del_list = "auto_test_TPE"
        for i in range(9):
            first_text = self.wait.until(
                EC.visibility_of_element_located((By.XPATH, '//tbody[@class="ant-table-tbody"]/tr[1]/td[2]'))).text
            if first_text == del_list:
                self.wait.until(EC.element_to_be_clickable(
                    (By.XPATH, '//tbody[@class="ant-table-tbody"]/tr[1]/td[10]/div'))).click()
                self.wait.until(EC.element_to_be_clickable((By.XPATH,
                                                            '//ul[@class="ant-dropdown-menu ant-dropdown-menu-root ant-dropdown-menu-vertical ant-dropdown-menu-light"]/li[4]/span[1]'))).click()
                self.wait.until(EC.element_to_be_clickable((By.XPATH,
                                                            '//div[@class="ant-modal-confirm-body-wrapper"]//button[@class="ant-btn ant-btn-primary"]'))).click()
                sleep(1)
            else:
                break
        self.driver.quit()

    def setUp(self) -> None:
        sleep(1)

    def tearDown(self) -> None:
        sleep(1)
        try:
            self.driver.find_element(By.XPATH,
                                     '//div[@class="ant-modal-mask"]/../div[2]/div[1]/div[2]/div[3]/div[1]/div[1]')
        except NoSuchElementException:
            pass
        else:
            self.wait.until(EC.element_to_be_clickable(
                (By.XPATH, '//div[@class="ant-modal-mask"]/../div[2]/div[1]/div[2]/div[3]/div[1]/div[1]'))).click()

    def test_required_para_empty(self):
        self.wait.until(EC.element_to_be_clickable((By.XPATH, '//div[@class="ant-space ant-space-horizontal ant-space-align-center ant-pro-table-list-toolbar-right"]//button[@class="ant-btn ant-btn-default"]'))).click()
        self.wait.until(EC.element_to_be_clickable(
            (By.XPATH, '//div[@class="ant-modal-mask"]/../div[2]/div[1]/div[2]/div[3]/div[1]/div[2]'))).click()
        error_dict = ["请输入Name", "请输入Data","请输入Algorithm"]
        error_info = self.wait.until(
            EC.visibility_of_element_located((By.XPATH, '//div[@class="ant-form-item-explain-error"]'))).text
        self.assertIn(error_info, error_dict)
        self.wait.until(EC.element_to_be_clickable(
            (By.XPATH, '//div[@class="ant-modal-mask"]/../div[2]/div[1]/div[2]/div[3]/div[1]/div[1]'))).click()

    def test_create_name_exsit(self):
        name = self.wait.until(EC.visibility_of_element_located((By.XPATH,'//tbody[@class="ant-table-tbody"]/tr[1]/td[2]'))).text
        self.wait.until(EC.element_to_be_clickable((By.XPATH,'//button[@class="ant-btn ant-btn-default"]'))).click()
        self.wait.until(EC.visibility_of_element_located((By.ID, "name"))).send_keys(Keys.CONTROL, "a")
        self.wait.until(EC.visibility_of_element_located((By.ID, "name"))).send_keys(Keys.BACKSPACE)
        self.wait.until(EC.visibility_of_element_located((By.ID, "name"))).send_keys(name)
        error_info = self.wait.until(EC.visibility_of_element_located((By.XPATH,'//div[@class="ant-form-item-explain ant-form-item-explain-connected"]/div'))).text
        self.assertIn(error_info,"Name名字重复!")
        self.wait.until(EC.element_to_be_clickable(
                (By.XPATH, '//div[@class="ant-modal-mask"]/../div[2]/div[1]/div[2]/div[3]/div[1]/div[1]'))).click()

    def test_Abnormal_input(self):
        self.wait.until(EC.element_to_be_clickable((By.XPATH,'//div[@class="ant-space ant-space-horizontal ant-space-align-center ant-pro-table-list-toolbar-right"]//button[@class="ant-btn ant-btn-default"]'))).click()
        self.wait.until(EC.visibility_of_element_located((By.ID, "name"))).send_keys(Keys.CONTROL, "a")
        self.wait.until(EC.visibility_of_element_located((By.ID, "name"))).send_keys(Keys.BACKSPACE)
        self.wait.until(EC.visibility_of_element_located((By.ID, "name"))).send_keys("@#@$")
        warn_info = self.wait.until(EC.visibility_of_element_located((By.XPATH, '//div[@class="ant-form-item-explain-error"]'))).text
        self.assertEqual(warn_info, "仅允许包含字母、数字、下划线!")
        self.wait.until(EC.element_to_be_clickable(
                (By.XPATH, '//div[@class="ant-modal-mask"]/../div[2]/div[1]/div[2]/div[3]/div[1]/div[1]'))).click()

    def test_rerun_required_para_empty(self):
        self.wait.until(
            EC.element_to_be_clickable((By.XPATH, '//tbody[@class="ant-table-tbody"]/tr[1]/td[10]/div'))).click()
        self.wait.until(EC.element_to_be_clickable(
            (By.XPATH, '//div[@class="ant-dropdown ant-dropdown-placement-bottomLeft "]/ul/li[3]/span[1]'))).click()
        self.wait.until(EC.visibility_of_element_located((By.ID, "name"))).send_keys(Keys.CONTROL, "a")
        self.wait.until(EC.visibility_of_element_located((By.ID, "name"))).send_keys(Keys.BACKSPACE)
        self.wait.until(EC.element_to_be_clickable(
            (By.XPATH, '//div[@class="ant-modal-mask"]/../div[2]/div[1]/div[2]/div[3]/div[1]/div[2]'))).click()
        error_dict = ["请输入Name", "请输入Data","请输入Algorithm"]
        error_info = self.wait.until(
            EC.visibility_of_element_located((By.XPATH, '//div[@class="ant-form-item-explain-error"]'))).text
        self.assertIn(error_info, error_dict)
        self.wait.until(EC.element_to_be_clickable(
            (By.XPATH, '//div[@class="ant-modal-mask"]/../div[2]/div[1]/div[2]/div[3]/div[1]/div[1]'))).click()

    def test_rerun_name_exsit(self):
        name = self.wait.until(EC.visibility_of_element_located((By.XPATH,'//tbody[@class="ant-table-tbody"]/tr[1]/td[2]'))).text
        self.wait.until(
            EC.element_to_be_clickable((By.XPATH, '//tbody[@class="ant-table-tbody"]/tr[1]/td[10]/div'))).click()
        self.wait.until(EC.element_to_be_clickable(
            (By.XPATH, '//div[@class="ant-dropdown ant-dropdown-placement-bottomLeft "]/ul/li[3]/span[1]'))).click()
        self.wait.until(EC.visibility_of_element_located((By.ID, "name"))).send_keys(Keys.CONTROL, "a")
        self.wait.until(EC.visibility_of_element_located((By.ID, "name"))).send_keys(Keys.BACKSPACE)
        self.wait.until(EC.visibility_of_element_located((By.ID, "name"))).send_keys(name)
        error_info = self.wait.until(EC.visibility_of_element_located((By.XPATH,'//div[@class="ant-form-item-explain ant-form-item-explain-connected"]/div'))).text
        self.assertIn(error_info,"Name名字重复!")
        self.wait.until(EC.element_to_be_clickable(
                (By.XPATH, '//div[@class="ant-modal-mask"]/../div[2]/div[1]/div[2]/div[3]/div[1]/div[1]'))).click()

    def test_rerun_Abnormal_input(self):
        self.wait.until(
            EC.element_to_be_clickable((By.XPATH, '//tbody[@class="ant-table-tbody"]/tr[1]/td[10]/div'))).click()
        self.wait.until(EC.element_to_be_clickable(
            (By.XPATH, '//div[@class="ant-dropdown ant-dropdown-placement-bottomLeft "]/ul/li[3]/span[1]'))).click()
        self.wait.until(EC.visibility_of_element_located((By.ID, "name"))).send_keys(Keys.CONTROL, "a")
        self.wait.until(EC.visibility_of_element_located((By.ID, "name"))).send_keys(Keys.BACKSPACE)
        self.wait.until(EC.visibility_of_element_located((By.ID, "name"))).send_keys("@#@$")
        warn_info = self.wait.until(EC.visibility_of_element_located((By.XPATH, '//div[@class="ant-form-item-explain-error"]'))).text
        self.assertEqual(warn_info, "仅允许包含字母、数字、下划线!")
        self.wait.until(EC.element_to_be_clickable(
                (By.XPATH, '//div[@class="ant-modal-mask"]/../div[2]/div[1]/div[2]/div[3]/div[1]/div[1]'))).click()
