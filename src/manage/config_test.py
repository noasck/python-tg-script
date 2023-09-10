"""Test config."""
import os

import pytest
from pydantic import ValidationError

from .config import _APP_ENV_PREFIX, APP_NAME, ConfigParser, _Config


@pytest.fixture()
def config_class() -> _Config:
    """Get config class with no instance created."""
    _Config._instance = None
    return _Config


def test_config_singleton(config_class):
    """Test positive case: singleton instance."""
    assert config_class() is config_class()


def test_ConfigParser_works():
    """Test positive case: serialize envs."""
    envs = {
        env.replace(_APP_ENV_PREFIX, ""): value
        for env, value in os.environ.items()
        if env.startswith(APP_NAME.upper())
    }
    parsed = ConfigParser(**envs)
    assert parsed
    assert parsed.bot_token
    assert parsed.secret_key


def test_ConfigParser_required_fields():
    """Test negative case: fields are not specified."""
    with pytest.raises(ValidationError) as err:
        ConfigParser()

    assert all(err["msg"] == "Field required" for err in err.value.errors())


def test_config_singleton(config_class):
    """Test positive case: singleton instance."""
    assert config_class() is config_class()


def test_ConfigParser_works():
    """Test positive case: serialize envs."""
    envs = {
        env.replace(_APP_ENV_PREFIX, ""): value
        for env, value in os.environ.items()
        if env.startswith(APP_NAME.upper())
    }
    parsed = ConfigParser(**envs)
    assert parsed
    assert parsed.bot_token
    assert parsed.secret_key


def test_ConfigParser_required_fields():
    """Test negative case: fields are not specified."""
    with pytest.raises(ValidationError) as err:
        ConfigParser()

    assert all(err["msg"] == "Field required" for err in err.value.errors())
